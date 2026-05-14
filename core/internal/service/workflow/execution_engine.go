package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

type ExecutionEngine struct{}

type engineNode struct {
	Id     string
	Type   string
	Config map[string]interface{}
}

type engineConnection struct {
	Id        string
	Source    string
	Target    string
	Condition string
}

func NewExecutionEngine() *ExecutionEngine {
	return &ExecutionEngine{}
}

func (e *ExecutionEngine) ExecuteWorkflow(ctx context.Context, workflowId int64, contactId int64, inputData map[string]interface{}) (*WorkflowExecution, error) {
	if workflowId <= 0 {
		return nil, gerror.New("workflow id is required")
	}
	if inputData == nil {
		inputData = map[string]interface{}{}
	}
	if err := e.ensureRuntimeTables(ctx); err != nil {
		return nil, err
	}
	workflow, err := WorkflowRepository().GetWorkflowById(ctx, workflowId)
	if err != nil {
		return nil, err
	}
	if workflow == nil {
		return nil, gerror.Newf("workflow %d not found", workflowId)
	}
	nodes, connections, err := e.loadGraph(ctx, workflowId)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, gerror.New("workflow has no nodes")
	}
	nodeById := make(map[string]*engineNode, len(nodes))
	incoming := make(map[string]int)
	outgoing := make(map[string][]*engineConnection)
	for _, node := range nodes {
		nodeById[node.Id] = node
	}
	for _, connection := range connections {
		incoming[connection.Target]++
		outgoing[connection.Source] = append(outgoing[connection.Source], connection)
	}
	start := findStartNode(nodes, incoming)
	if start == nil {
		return nil, gerror.New("workflow start node not found")
	}
	now := time.Now().Unix()
	execution := &WorkflowExecution{WorkflowId: workflowId, Version: workflow.Version, Status: 1, Trigger: gconv.String(inputData["trigger"]), StartedAt: now, CreatedAt: now, UpdatedAt: now}
	executionId, err := WorkflowRepository().RecordExecution(ctx, execution)
	if err != nil {
		return nil, err
	}
	execution.Id = executionId
	contactEmail := firstString(inputData, "contact_email", "email")
	if _, err := g.DB().Model("workflow_execution").Ctx(ctx).Where("id", executionId).Data(g.Map{
		"contact_email":   contactEmail,
		"contact_id_text": gconv.String(contactId),
	}).Update(); err != nil {
		return nil, err
	}
	status := "completed"
	errMessage := ""
	visited := map[string]bool{}
	current := start
	for current != nil {
		if visited[current.Id] {
			status = "failed"
			errMessage = fmt.Sprintf("cycle detected at node %s", current.Id)
			break
		}
		visited[current.Id] = true
		result, nodeErr := e.executeNode(ctx, executionId, workflowId, contactId, current, inputData)
		if nodeErr != nil {
			status = "failed"
			errMessage = nodeErr.Error()
			break
		}
		current = chooseNextNode(current, result, outgoing[current.Id], nodeById)
	}
	completedAt := time.Now().Unix()
	if status == "completed" {
		execution.Status = 2
	} else {
		execution.Status = 3
		execution.Error = errMessage
	}
	execution.CompletedAt = completedAt
	execution.Duration = completedAt - execution.StartedAt
	_, err = g.DB().Model("workflow_execution").Ctx(ctx).Where("id", executionId).Data(g.Map{"status": status, "completed_at": completedAt, "duration": execution.Duration, "error_message": errMessage, "updated_at": completedAt}).Update()
	if err != nil {
		return nil, err
	}
	return execution, nil
}

func (e *ExecutionEngine) loadGraph(ctx context.Context, workflowId int64) ([]*engineNode, []*engineConnection, error) {
	nodeRecords, err := g.DB().Model("workflow_node").Ctx(ctx).Where("workflow_id", workflowId).Order("created_at asc").All()
	if err != nil {
		return nil, nil, err
	}
	connectionRecords, err := g.DB().Model("workflow_connection").Ctx(ctx).Where("workflow_id", workflowId).Order("created_at asc").All()
	if err != nil {
		return nil, nil, err
	}
	nodes := make([]*engineNode, 0, len(nodeRecords))
	for _, record := range nodeRecords {
		config := map[string]interface{}{}
		if raw := record["config"].String(); raw != "" {
			_ = json.Unmarshal([]byte(raw), &config)
		}
		nodes = append(nodes, &engineNode{Id: record["id"].String(), Type: record["type"].String(), Config: config})
	}
	connections := make([]*engineConnection, 0, len(connectionRecords))
	for _, record := range connectionRecords {
		connections = append(connections, &engineConnection{Id: record["id"].String(), Source: record["source"].String(), Target: record["target"].String(), Condition: record["condition"].String()})
	}
	return nodes, connections, nil
}

func (e *ExecutionEngine) executeNode(ctx context.Context, executionId, workflowId, contactId int64, node *engineNode, inputData map[string]interface{}) (bool, error) {
	if executionId <= 0 {
		return false, gerror.New("execution id is required before writing workflow execution logs")
	}
	exists, err := g.DB().Model("workflow_execution").Ctx(ctx).Where("id", executionId).Exist()
	if err != nil {
		return false, err
	}
	if !exists {
		return false, gerror.Newf("workflow execution %d does not exist", executionId)
	}
	startedAt := time.Now().Unix()
	action := node.Type
	status := "success"
	message := ""
	result := true
	details := map[string]interface{}{"config": node.Config}
	switch node.Type {
	case "trigger":
		action = "trigger"
		message = "Trigger passed"
	case "send-email", "email":
		action = "email_sent"
		details["template_id"] = gconv.String(node.Config["templateId"])
		details["to"] = firstString(inputData, "contact_email", "email")
		message = fmt.Sprintf("Email would be sent to %s using template %s", details["to"], details["template_id"])
		fmt.Println(message)
	case "delay":
		action = "delay"
		message = fmt.Sprintf("Delay would wait %v %v", node.Config["duration"], node.Config["unit"])
	case "condition":
		action = "condition"
		result = evaluateConditionGroup(node.Config, inputData)
		details["result"] = result
		message = fmt.Sprintf("Condition evaluated to %v", result)
	case "action":
		action = gconv.String(node.Config["actionType"])
		if action == "" {
			action = "action"
		}
		message = fmt.Sprintf("Action would execute: %s %v", action, node.Config["value"])
	case "split":
		action = "split"
		message = "Split selected first branch for demo execution"
	default:
		message = fmt.Sprintf("Node %s processed", node.Type)
	}
	finishedAt := time.Now().Unix()
	detailsJSON, _ := json.Marshal(details)
	_, err = g.DB().Model("workflow_execution_log").Ctx(ctx).Data(g.Map{"workflow_id": workflowId, "execution_id": executionId, "contact_id": gconv.String(contactId), "node_id": node.Id, "node_type": node.Type, "status": status, "level": "info", "message": message, "started_at": startedAt, "finished_at": finishedAt, "timestamp": finishedAt}).Insert()
	if err != nil {
		return result, err
	}
	_, _ = g.DB().Model("workflow_log").Ctx(ctx).Data(g.Map{"execution_id": gconv.String(executionId), "node_id": node.Id, "node_type": node.Type, "action": action, "status": status, "details": string(detailsJSON), "created_at": time.Now()}).Insert()
	return result, nil
}

func (e *ExecutionEngine) ensureRuntimeTables(ctx context.Context) error {
	if err := WorkflowEditor().ensureEditorTables(ctx, g.DB()); err != nil {
		return err
	}
	if err := GetWorkflowService().EnsureExecutionLogTable(ctx); err != nil {
		return err
	}
	queries := []string{
		`ALTER TABLE workflow_execution ADD COLUMN IF NOT EXISTS contact_email VARCHAR(320) NOT NULL DEFAULT ''`,
		`ALTER TABLE workflow_execution ADD COLUMN IF NOT EXISTS contact_id_text VARCHAR(320) NOT NULL DEFAULT ''`,
		`ALTER TABLE workflow_execution ADD COLUMN IF NOT EXISTS duration BIGINT DEFAULT 0`,
		`ALTER TABLE workflow_execution ADD COLUMN IF NOT EXISTS error_message TEXT DEFAULT ''`,
		`ALTER TABLE workflow_execution ADD COLUMN IF NOT EXISTS updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())`,
		`ALTER TABLE workflow_log ADD COLUMN IF NOT EXISTS details TEXT NOT NULL DEFAULT ''`,
	}
	for _, query := range queries {
		if _, err := g.DB().Exec(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func findStartNode(nodes []*engineNode, incoming map[string]int) *engineNode {
	for _, node := range nodes {
		if node.Type == "trigger" {
			return node
		}
	}
	for _, node := range nodes {
		if incoming[node.Id] == 0 {
			return node
		}
	}
	return nodes[0]
}

func chooseNextNode(node *engineNode, conditionResult bool, connections []*engineConnection, nodeById map[string]*engineNode) *engineNode {
	if len(connections) == 0 {
		return nil
	}
	if node.Type == "condition" {
		want := "false"
		if conditionResult {
			want = "true"
		}
		for _, connection := range connections {
			label := strings.ToLower(strings.TrimSpace(connection.Condition))
			if label == want || label == "if "+want || label == "yes" && conditionResult || label == "no" && !conditionResult {
				return nodeById[connection.Target]
			}
		}
	}
	return nodeById[connections[0].Target]
}

func evaluateConditionGroup(group map[string]interface{}, inputData map[string]interface{}) bool {
	logic := strings.ToUpper(gconv.String(group["logic"]))
	if logic == "" {
		logic = "AND"
	}
	results := make([]bool, 0)
	if conditions, ok := group["conditions"].([]interface{}); ok {
		for _, item := range conditions {
			if condition, ok := item.(map[string]interface{}); ok {
				results = append(results, evaluateCondition(condition, inputData))
			}
		}
	}
	if groups, ok := group["groups"].([]interface{}); ok {
		for _, item := range groups {
			if child, ok := item.(map[string]interface{}); ok {
				results = append(results, evaluateConditionGroup(child, inputData))
			}
		}
	}
	if len(results) == 0 {
		return true
	}
	if logic == "OR" {
		for _, result := range results {
			if result {
				return true
			}
		}
		return false
	}
	for _, result := range results {
		if !result {
			return false
		}
	}
	return true
}

func evaluateCondition(condition map[string]interface{}, inputData map[string]interface{}) bool {
	field := gconv.String(condition["field"])
	operator := gconv.String(condition["operator"])
	expected := gconv.String(condition["value"])
	actual := gconv.String(inputData[field])
	switch operator {
	case "not_equals":
		return actual != expected
	case "contains":
		return strings.Contains(actual, expected)
	case "greater_than":
		return gconv.Float64(actual) > gconv.Float64(expected)
	case "less_than":
		return gconv.Float64(actual) < gconv.Float64(expected)
	case "between":
		value := gconv.Float64(actual)
		return value >= gconv.Float64(expected) && value <= gconv.Float64(condition["valueTo"])
	default:
		return actual == expected
	}
}

func firstString(data map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if value := gconv.String(data[key]); value != "" {
			return value
		}
	}
	return ""
}
