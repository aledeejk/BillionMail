package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"billionmail-core/internal/service/workflow/queue"

	contactSvc "billionmail-core/internal/service/contact"
	emailTplSvc "billionmail-core/internal/service/email_template"
	mailSvc "billionmail-core/internal/service/mail_service"

	"github.com/go-redsync/redsync/v4"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

type ExecutionEngine struct {
	q *queue.RedisQueue
}

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

func NewExecutionEngine(q *queue.RedisQueue) *ExecutionEngine {
	return &ExecutionEngine{q: q}
}

func (e *ExecutionEngine) HasQueue() bool {
	return e.q != nil
}

func (e *ExecutionEngine) DispatchExecution(ctx context.Context, workflowId int64, contactId int64, inputData map[string]interface{}) error {
	if e.q == nil {
		return gerror.New("redis queue is not initialised: cannot dispatch async execution")
	}
	idempotencyKey := firstString(inputData, "idempotency_key")
	if idempotencyKey == "" {
		email := firstString(inputData, "contact_email", "email")
		idempotencyKey = fmt.Sprintf("%d:%s:%d", workflowId, email, time.Now().UnixNano())
	}
	task := queue.ExecutionTask{
		WorkflowId:     workflowId,
		ContactId:      contactId,
		InputData:      inputData,
		IdempotencyKey: idempotencyKey,
	}
	return e.q.Enqueue(ctx, task)
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

	idempotencyKey := firstString(inputData, "idempotency_key")

	var mu *redsync.Mutex
	if e.q != nil {
		var lockErr error
		mu, lockErr = e.q.AcquireWorkflowLockWithKey(workflowId, idempotencyKey)
		if lockErr != nil {
			return nil, gerror.Newf("workflow %d execution %q is already running on another instance", workflowId, idempotencyKey)
		}
		defer func() { _, _ = mu.Unlock() }()
	}

	if idempotencyKey != "" {
		var existing struct {
			Id int64 `json:"id"`
		}
		_ = g.DB().Model("workflow_execution").Ctx(ctx).
			Where("workflow_id", workflowId).
			Where("idempotency_key", idempotencyKey).
			Fields("id").
			Scan(&existing)
		if existing.Id > 0 {
			execution := &WorkflowExecution{Id: existing.Id, WorkflowId: workflowId}
			return execution, nil
		}
	}

	workflow, err := WorkflowRepository().GetWorkflowById(ctx, workflowId)
	if err != nil {
		return nil, err
	}
	if workflow == nil {
		return nil, gerror.Newf("workflow %d not found", workflowId)
	}
	if workflow.Status == 0 && gconv.String(inputData["force_execute"]) != "true" {
		return nil, gerror.Newf("workflow %d is inactive, cannot execute", workflowId)
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
	execution := &WorkflowExecution{
		WorkflowId: workflowId,
		Version:    workflow.Version,
		Status:     1,
		Trigger:    gconv.String(inputData["trigger"]),
		StartedAt:  now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	executionId, err := WorkflowRepository().RecordExecution(ctx, execution)
	if err != nil {
		return nil, err
	}
	execution.Id = executionId

	contextJSON, _ := json.Marshal(inputData)
	contactEmail := firstString(inputData, "contact_email", "email")
	if _, err := g.DB().Model("workflow_execution").Ctx(ctx).Where("id", executionId).Data(g.Map{
		"contact_email":   contactEmail,
		"contact_id_text": gconv.String(contactId),
		"idempotency_key": idempotencyKey,
		"current_node_id": start.Id,
		"context":         string(contextJSON),
		"retry_count":     0,
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

		if _, updateErr := g.DB().Model("workflow_execution").Ctx(ctx).
			Where("id", executionId).
			Data(g.Map{"current_node_id": current.Id}).Update(); updateErr != nil {
			fmt.Printf("[engine] warn: could not update current_node_id: %v\n", updateErr)
		}

		result, nodeErr := e.executeNode(ctx, executionId, workflowId, contactId, current, inputData)
		if nodeErr != nil {
			status = "failed"
			errMessage = nodeErr.Error()

			g.DB().Model("workflow_execution").Ctx(ctx).Where("id", executionId).
				Data(g.Map{"retry_count": g.DB().Raw("retry_count + 1")}).Update()
			break
		}
		if current.Type == "delay" {
			status = "paused"
			nextNode := chooseNextNode(current, true, outgoing[current.Id], nodeById)
			nextId := ""
			if nextNode != nil {
				nextId = nextNode.Id
			}
			g.DB().Model("workflow_execution").Ctx(ctx).Where("id", executionId).
				Data(g.Map{"current_node_id": nextId}).Update()
			break
		}
		current = chooseNextNode(current, result, outgoing[current.Id], nodeById)
	}

	completedAt := time.Now().Unix()
	if status == "completed" {
		execution.Status = 2
	} else if status == "paused" {
		execution.Status = 4
	} else {
		execution.Status = 3
		execution.Error = errMessage
	}
	execution.CompletedAt = completedAt
	execution.Duration = completedAt - execution.StartedAt
	if status == "paused" {
		_, err = g.DB().Model("workflow_execution").Ctx(ctx).Where("id", executionId).Data(g.Map{
			"status":     status,
			"updated_at": completedAt,
		}).Update()
	} else {
		_, err = g.DB().Model("workflow_execution").Ctx(ctx).Where("id", executionId).Data(g.Map{
			"status":          status,
			"completed_at":    completedAt,
			"duration":        execution.Duration,
			"error_message":   errMessage,
			"updated_at":      completedAt,
			"current_node_id": "",
		}).Update()
	}
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
		alreadySent, _ := g.DB().Model("workflow_log").Ctx(ctx).
			Where("execution_id", gconv.String(executionId)).
			Where("node_id", node.Id).
			Where("action", "email_sent").
			Where("status", "success").
			Count()
		if alreadySent > 0 {
			message = fmt.Sprintf("send-email: skipped duplicate send (node %s already succeeded in execution %d)", node.Id, executionId)
			break
		}
		recipientEmail := firstString(inputData, "contact_email", "email")
		templateId := gconv.Int(node.Config["templateId"])
		senderEmail := gconv.String(node.Config["sender"])
		details["template_id"] = gconv.String(templateId)
		details["to"] = recipientEmail
		if recipientEmail == "" {
			status = "failed"
			message = "send-email: recipient email is empty"
			result = false
			break
		}
		if templateId <= 0 {
			status = "failed"
			message = "send-email: templateId is not configured"
			result = false
			break
		}
		emailTpl, tplErr := emailTplSvc.GetTemplatesByID(ctx, templateId)
		if tplErr != nil || emailTpl == nil {
			status = "failed"
			message = fmt.Sprintf("send-email: template %d not found: %v", templateId, tplErr)
			result = false
			break
		}
		sender, senderErr := mailSvc.NewWorkflowEmailSender(ctx, senderEmail)
		if senderErr != nil {
			status = "failed"
			message = fmt.Sprintf("send-email: could not create sender %q: %v", senderEmail, senderErr)
			result = false
			break
		}
		defer sender.Close()
		subject := gconv.String(node.Config["subject"])
		if subject == "" {
			subject = emailTpl.TempName
		}
		mailMsg := mailSvc.NewMessage(subject, emailTpl.Content)
		if sendErr := sender.Send(mailMsg, []string{recipientEmail}); sendErr != nil {
			status = "failed"
			message = fmt.Sprintf("send-email: SMTP error to %s: %v", recipientEmail, sendErr)
			result = false
		} else {
			message = fmt.Sprintf("Email sent to %s via template %d (%s)", recipientEmail, templateId, emailTpl.TempName)
		}
	case "delay":
		action = "delay"
		duration := gconv.Int64(node.Config["duration"])
		unit := gconv.String(node.Config["unit"])
		if duration <= 0 {
			duration = 1
		}
		if unit == "" {
			unit = "days"
		}
		message = fmt.Sprintf("Execution paused: delay %d %s (will resume after delay)", duration, unit)
		result = false
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
		contactEmail := firstString(inputData, "contact_email", "email")
		switch action {
		case "add-tag":
			tagName := gconv.String(node.Config["value"])
			groupId := gconv.Int(node.Config["groupId"])
			if contactEmail != "" && tagName != "" {
				if tagErr := contactSvc.AddTagToContact(ctx, contactEmail, groupId, tagName); tagErr != nil {
					status = "failed"
					message = fmt.Sprintf("add-tag: %v", tagErr)
					result = false
				} else {
					message = fmt.Sprintf("Tag %q added to contact %s", tagName, contactEmail)
				}
			} else {
				message = fmt.Sprintf("add-tag skipped: email=%q tag=%q", contactEmail, tagName)
			}
		case "remove-tag":
			tagName := gconv.String(node.Config["value"])
			groupId := gconv.Int(node.Config["groupId"])
			if contactEmail != "" && tagName != "" {
				if tagErr := contactSvc.RemoveTagFromContact(ctx, contactEmail, groupId, tagName); tagErr != nil {
					status = "failed"
					message = fmt.Sprintf("remove-tag: %v", tagErr)
					result = false
				} else {
					message = fmt.Sprintf("Tag %q removed from contact %s", tagName, contactEmail)
				}
			} else {
				message = fmt.Sprintf("remove-tag skipped: email=%q tag=%q", contactEmail, tagName)
			}
		case "move-to-group":
			targetGroupId := gconv.Int(node.Config["value"])
			if contactEmail != "" && targetGroupId > 0 {
				if moveErr := contactSvc.MoveContactToGroup(ctx, contactEmail, targetGroupId); moveErr != nil {
					status = "failed"
					message = fmt.Sprintf("move-to-group: %v", moveErr)
					result = false
				} else {
					message = fmt.Sprintf("Contact %s moved to group %d", contactEmail, targetGroupId)
				}
			} else {
				message = fmt.Sprintf("move-to-group skipped: email=%q groupId=%d", contactEmail, targetGroupId)
			}
		default:
			message = fmt.Sprintf("Action %s executed with value %v", action, node.Config["value"])
		}
	case "split":
		action = "split"
		branchA := gconv.Float64(node.Config["branchA"])
		branchB := gconv.Float64(node.Config["branchB"])
		if branchA <= 0 && branchB <= 0 {
			branchA = 50
			branchB = 50
		}
		total := branchA + branchB
		if total <= 0 {
			total = 100
		}
		contactSeed := firstString(inputData, "contact_email", "email", "contact_id")
		var hashVal uint32
		for _, ch := range contactSeed + node.Id {
			hashVal = hashVal*31 + uint32(ch)
		}
		norm := float64(hashVal%10000) / 100.0
		threshold := branchA / total * 100.0
		if norm < threshold {
			result = true
			message = fmt.Sprintf("Split: contact %q → branch A (%.0f%%) [hash=%.2f threshold=%.2f]", contactSeed, branchA, norm, threshold)
			details["branch"] = "A"
		} else {
			result = false
			message = fmt.Sprintf("Split: contact %q → branch B (%.0f%%) [hash=%.2f threshold=%.2f]", contactSeed, branchB, norm, threshold)
			details["branch"] = "B"
		}
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
		`ALTER TABLE workflow_execution ADD COLUMN IF NOT EXISTS current_node_id VARCHAR(64) NOT NULL DEFAULT ''`,
		`ALTER TABLE workflow_execution ADD COLUMN IF NOT EXISTS context JSONB NOT NULL DEFAULT '{}'`,
		`ALTER TABLE workflow_execution ADD COLUMN IF NOT EXISTS retry_count INT NOT NULL DEFAULT 0`,
		`ALTER TABLE workflow_execution ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(128) NOT NULL DEFAULT ''`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_workflow_execution_idempotency ON workflow_execution (workflow_id, idempotency_key) WHERE idempotency_key <> ''`,
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
