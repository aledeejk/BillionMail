package workflow

import (
	"context"
	"encoding/json"

	"github.com/gogf/gf/v2/frame/g"
)

type IWorkflowRepository interface {
	CreateWorkflow(ctx context.Context, workflow *Workflow) (int64, error)
	UpdateWorkflow(ctx context.Context, workflow *Workflow) error
	DeleteWorkflow(ctx context.Context, workflowId int64) error
	GetWorkflowById(ctx context.Context, workflowId int64) (*Workflow, error)
	ListWorkflows(ctx context.Context, page, pageSize int, keyword string, status int) ([]*Workflow, int, error)
	CreateWorkflowVersion(ctx context.Context, version *WorkflowVersion) (int64, error)
	GetWorkflowVersions(ctx context.Context, workflowId int64) ([]*WorkflowVersion, error)
	GetWorkflowVersionById(ctx context.Context, versionId int64) (*WorkflowVersion, error)
	RecordExecution(ctx context.Context, execution *WorkflowExecution) (int64, error)
	ListExecutions(ctx context.Context, workflowId int64, page, pageSize int) ([]*WorkflowExecution, int, error)
	CreateLog(ctx context.Context, log *WorkflowLog) (int64, error)
	ListLogs(ctx context.Context, workflowId int64, page, pageSize int) ([]*WorkflowLog, int, error)
}

type workflowRepository struct{}

var (
	workflowRepoInstance IWorkflowRepository
)

func WorkflowRepository() IWorkflowRepository {
	if workflowRepoInstance == nil {
		workflowRepoInstance = newWorkflowRepository()
	}
	return workflowRepoInstance
}

func newWorkflowRepository() *workflowRepository {
	return &workflowRepository{}
}

func (r *workflowRepository) CreateWorkflow(ctx context.Context, workflow *Workflow) (int64, error) {
	data := g.Map{
		"name":        workflow.Name,
		"description": workflow.Description,
		"status":      workflow.Status,
		"version":     workflow.Version,
		"trigger":     workflow.Trigger,
		"created_at":  workflow.CreatedAt,
		"updated_at":  workflow.UpdatedAt,
	}

	if workflow.Nodes != nil {
		nodes, err := json.Marshal(workflow.Nodes)
		if err != nil {
			return 0, err
		}
		data["nodes"] = string(nodes)
	}
	if workflow.Connections != nil {
		connections, err := json.Marshal(workflow.Connections)
		if err != nil {
			return 0, err
		}
		data["connections"] = string(connections)
	}
	if workflow.Metadata != nil {
		metadata, err := json.Marshal(workflow.Metadata)
		if err != nil {
			return 0, err
		}
		data["metadata"] = string(metadata)
	}

	result, err := g.DB().Model("bm_workflows").Ctx(ctx).Data(data).Insert()
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *workflowRepository) UpdateWorkflow(ctx context.Context, workflow *Workflow) error {
	data := g.Map{
		"name":        workflow.Name,
		"description": workflow.Description,
		"status":      workflow.Status,
		"version":     workflow.Version,
		"trigger":     workflow.Trigger,
		"updated_at":  workflow.UpdatedAt,
	}

	if workflow.Nodes != nil {
		nodes, err := json.Marshal(workflow.Nodes)
		if err != nil {
			return err
		}
		data["nodes"] = string(nodes)
	}
	if workflow.Connections != nil {
		connections, err := json.Marshal(workflow.Connections)
		if err != nil {
			return err
		}
		data["connections"] = string(connections)
	}
	if workflow.Metadata != nil {
		metadata, err := json.Marshal(workflow.Metadata)
		if err != nil {
			return err
		}
		data["metadata"] = string(metadata)
	}

	_, err := g.DB().Model("bm_workflows").Ctx(ctx).Where("id", workflow.Id).Data(data).Update()
	return err
}

func (r *workflowRepository) DeleteWorkflow(ctx context.Context, workflowId int64) error {
	_, err := g.DB().Model("bm_workflows").Ctx(ctx).Where("id", workflowId).Delete()
	if err != nil {
		return err
	}
	_, _ = g.DB().Model("bm_workflow_versions").Ctx(ctx).Where("workflow_id", workflowId).Delete()
	_, _ = g.DB().Model("bm_workflow_executions").Ctx(ctx).Where("workflow_id", workflowId).Delete()
	_, _ = g.DB().Model("bm_workflow_logs").Ctx(ctx).Where("workflow_id", workflowId).Delete()
	return nil
}

func (r *workflowRepository) GetWorkflowById(ctx context.Context, workflowId int64) (*Workflow, error) {
	var record struct {
		Id          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      int    `json:"status"`
		Version     int    `json:"version"`
		Trigger     string `json:"trigger"`
		Nodes       string `json:"nodes"`
		Connections string `json:"connections"`
		Metadata    string `json:"metadata"`
		CreatedAt   int64  `json:"created_at"`
		UpdatedAt   int64  `json:"updated_at"`
	}

	err := g.DB().Model("bm_workflows").Ctx(ctx).Where("id", workflowId).Scan(&record)
	if err != nil {
		return nil, err
	}
	if record.Id == 0 {
		return nil, nil
	}

	workflow := &Workflow{
		Id:          record.Id,
		Name:        record.Name,
		Description: record.Description,
		Status:      record.Status,
		Version:     record.Version,
		Trigger:     record.Trigger,
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
	}

	if record.Nodes != "" {
		_ = json.Unmarshal([]byte(record.Nodes), &workflow.Nodes)
	}
	if record.Connections != "" {
		_ = json.Unmarshal([]byte(record.Connections), &workflow.Connections)
	}
	if record.Metadata != "" {
		_ = json.Unmarshal([]byte(record.Metadata), &workflow.Metadata)
	}

	return workflow, nil
}

func (r *workflowRepository) ListWorkflows(ctx context.Context, page, pageSize int, keyword string, status int) ([]*Workflow, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	model := g.DB().Model("bm_workflows").Ctx(ctx).Safe()
	if keyword != "" {
		model = model.WhereLike("name", "%"+keyword+"%").WhereOrLike("description", "%"+keyword+"%")
	}
	if status != -1 {
		model = model.Where("status", status)
	}

	total, err := model.Count()
	if err != nil {
		return nil, 0, err
	}

	var records []struct {
		Id          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      int    `json:"status"`
		Version     int    `json:"version"`
		Trigger     string `json:"trigger"`
		Nodes       string `json:"nodes"`
		Connections string `json:"connections"`
		Metadata    string `json:"metadata"`
		CreatedAt   int64  `json:"created_at"`
		UpdatedAt   int64  `json:"updated_at"`
	}
	
	err = model.Page(page, pageSize).Order("updated_at desc").Scan(&records)
	if err != nil {
		return nil, 0, err
	}

	workflows := make([]*Workflow, 0, len(records))
	for _, record := range records {
		workflow := &Workflow{
			Id:          record.Id,
			Name:        record.Name,
			Description: record.Description,
			Status:      record.Status,
			Version:     record.Version,
			Trigger:     record.Trigger,
			CreatedAt:   record.CreatedAt,
			UpdatedAt:   record.UpdatedAt,
		}
		if record.Nodes != "" {
			_ = json.Unmarshal([]byte(record.Nodes), &workflow.Nodes)
		}
		if record.Connections != "" {
			_ = json.Unmarshal([]byte(record.Connections), &workflow.Connections)
		}
		if record.Metadata != "" {
			_ = json.Unmarshal([]byte(record.Metadata), &workflow.Metadata)
		}
		workflows = append(workflows, workflow)
	}

	return workflows, int(total), nil
}

func (r *workflowRepository) CreateWorkflowVersion(ctx context.Context, version *WorkflowVersion) (int64, error) {
	data := g.Map{
		"workflow_id": version.WorkflowId,
		"version":     version.Version,
		"name":        version.Name,
		"description": version.Description,
		"status":      version.Status,
		"trigger":     version.Trigger,
		"definition":  version.Definition,
		"created_at":  version.CreatedAt,
	}

	result, err := g.DB().Model("bm_workflow_versions").Ctx(ctx).Data(data).Insert()
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *workflowRepository) GetWorkflowVersions(ctx context.Context, workflowId int64) ([]*WorkflowVersion, error) {
	var versions []*WorkflowVersion
	err := g.DB().Model("bm_workflow_versions").Ctx(ctx).Where("workflow_id", workflowId).Order("version desc").Scan(&versions)
	return versions, err
}

func (r *workflowRepository) GetWorkflowVersionById(ctx context.Context, versionId int64) (*WorkflowVersion, error) {
	var version WorkflowVersion
	err := g.DB().Model("bm_workflow_versions").Ctx(ctx).Where("id", versionId).Scan(&version)
	if err != nil {
		return nil, err
	}
	if version.Id == 0 {
		return nil, nil
	}
	return &version, nil
}

func (r *workflowRepository) RecordExecution(ctx context.Context, execution *WorkflowExecution) (int64, error) {
	data := g.Map{
		"workflow_id":  execution.WorkflowId,
		"version":      execution.Version,
		"status":       execution.Status,
		"trigger":      execution.Trigger,
		"started_at":   execution.StartedAt,
		"completed_at": execution.CompletedAt,
		"duration":     execution.Duration,
		"result":       execution.Result,
		"error":        execution.Error,
		"created_at":   execution.CreatedAt,
		"updated_at":   execution.UpdatedAt,
	}

	result, err := g.DB().Model("bm_workflow_executions").Ctx(ctx).Data(data).Insert()
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *workflowRepository) ListExecutions(ctx context.Context, workflowId int64, page, pageSize int) ([]*WorkflowExecution, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	model := g.DB().Model("bm_workflow_executions").Ctx(ctx).Where("workflow_id", workflowId).Safe()
	total, err := model.Count()
	if err != nil {
		return nil, 0, err
	}

	var executions []*WorkflowExecution
	err = model.Page(page, pageSize).Order("started_at desc").Scan(&executions)
	return executions, int(total), err
}

func (r *workflowRepository) CreateLog(ctx context.Context, log *WorkflowLog) (int64, error) {
	data := g.Map{
		"workflow_id":  log.WorkflowId,
		"execution_id": log.ExecutionId,
		"level":        log.Level,
		"message":      log.Message,
		"created_at":   log.CreatedAt,
	}

	result, err := g.DB().Model("bm_workflow_logs").Ctx(ctx).Data(data).Insert()
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *workflowRepository) ListLogs(ctx context.Context, workflowId int64, page, pageSize int) ([]*WorkflowLog, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	model := g.DB().Model("bm_workflow_logs").Ctx(ctx).Where("workflow_id", workflowId).Safe()
	total, err := model.Count()
	if err != nil {
		return nil, 0, err
	}

	var logs []*WorkflowLog
	err = model.Page(page, pageSize).Order("created_at desc").Scan(&logs)
	return logs, int(total), err
}
