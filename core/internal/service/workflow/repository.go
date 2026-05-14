package workflow

import (
	"context"

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
		"created_at":  workflow.CreatedAt,
		"updated_at":  workflow.UpdatedAt,
	}

	result, err := g.DB().Model("workflow").Ctx(ctx).Data(data).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *workflowRepository) UpdateWorkflow(ctx context.Context, workflow *Workflow) error {
	data := g.Map{
		"name":        workflow.Name,
		"description": workflow.Description,
		"status":      workflow.Status,
		"version":     workflow.Version,
		"updated_at":  workflow.UpdatedAt,
	}

	_, err := g.DB().Model("workflow").Ctx(ctx).Where("id", workflow.Id).Data(data).Update()
	return err
}

func (r *workflowRepository) DeleteWorkflow(ctx context.Context, workflowId int64) error {
	_, err := g.DB().Model("workflow").Ctx(ctx).Where("id", workflowId).Delete()
	if err != nil {
		return err
	}
	return nil
}

func (r *workflowRepository) GetWorkflowById(ctx context.Context, workflowId int64) (*Workflow, error) {
	var record struct {
		Id          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      int    `json:"status"`
		Version     int    `json:"version"`
		CreatedAt   int64  `json:"created_at"`
		UpdatedAt   int64  `json:"updated_at"`
	}

	err := g.DB().Model("workflow").Ctx(ctx).Where("id", workflowId).Scan(&record)
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
		CreatedAt:   record.CreatedAt,
		UpdatedAt:   record.UpdatedAt,
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
	if pageSize > 50 {
		pageSize = 50
	}

	model := g.DB().Model("workflow").Ctx(ctx).Safe()
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
		CreatedAt   int64  `json:"created_at"`
		UpdatedAt   int64  `json:"updated_at"`
	}

	err = model.
		Fields("id, name, description, status, version, created_at, updated_at").
		Page(page, pageSize).
		Order("id desc").
		Scan(&records)
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
			CreatedAt:   record.CreatedAt,
			UpdatedAt:   record.UpdatedAt,
		}
		workflows = append(workflows, workflow)
	}

	return workflows, int(total), nil
}

func (r *workflowRepository) CreateWorkflowVersion(ctx context.Context, version *WorkflowVersion) (int64, error) {
	data := g.Map{
		"workflow_id": version.WorkflowId,
		"version":     version.Version,
		"content":     version.Definition,
		"created_at":  version.CreatedAt,
	}

	result, err := g.DB().Model("workflow_version").Ctx(ctx).Data(data).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *workflowRepository) GetWorkflowVersions(ctx context.Context, workflowId int64) ([]*WorkflowVersion, error) {
	workflow, err := r.GetWorkflowById(ctx, workflowId)
	if err != nil {
		return nil, err
	}
	currentVersion := 0
	if workflow != nil {
		currentVersion = workflow.Version
	}

	var versions []*WorkflowVersion
	err = g.DB().Model("workflow_version").Ctx(ctx).
		Fields("version, created_at, 0 AS status").
		Where("workflow_id", workflowId).
		Order("version desc").
		Limit(50).
		Scan(&versions)
	if err != nil {
		return nil, err
	}

	for _, version := range versions {
		if version.Version == currentVersion {
			version.Status = 1
		}
	}

	return versions, nil
}

func (r *workflowRepository) GetWorkflowVersionById(ctx context.Context, versionId int64) (*WorkflowVersion, error) {
	var version WorkflowVersion
	err := g.DB().Model("workflow_version").Ctx(ctx).Fields("id, workflow_id, version, content AS definition, created_at").Where("id", versionId).Scan(&version)
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
		"workflow_id":   execution.WorkflowId,
		"status":        executionStatusText(execution.Status),
		"started_at":    execution.StartedAt,
		"completed_at":  execution.CompletedAt,
		"error_message": execution.Error,
	}

	result, err := g.DB().Model("workflow_execution").Ctx(ctx).Data(data).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *workflowRepository) ListExecutions(ctx context.Context, workflowId int64, page, pageSize int) ([]*WorkflowExecution, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	model := g.DB().Model("workflow_execution").Ctx(ctx).Where("workflow_id", workflowId).Safe()
	total, err := model.Count()
	if err != nil {
		return nil, 0, err
	}

	var records []struct {
		Id           int64  `json:"id"`
		WorkflowId   int64  `json:"workflow_id"`
		Status       string `json:"status"`
		StartedAt    int64  `json:"started_at"`
		CompletedAt  int64  `json:"completed_at"`
		ErrorMessage string `json:"error_message"`
	}
	err = model.Page(page, pageSize).Order("started_at desc").Scan(&records)
	if err != nil {
		return nil, 0, err
	}
	executions := make([]*WorkflowExecution, 0, len(records))
	for _, record := range records {
		executions = append(executions, &WorkflowExecution{
			Id:          record.Id,
			WorkflowId:  record.WorkflowId,
			Status:      executionStatusCode(record.Status),
			StartedAt:   record.StartedAt,
			CompletedAt: record.CompletedAt,
			Error:       record.ErrorMessage,
		})
	}
	return executions, int(total), err
}

func (r *workflowRepository) CreateLog(ctx context.Context, log *WorkflowLog) (int64, error) {
	data := g.Map{
		"workflow_id":  log.WorkflowId,
		"execution_id": log.ExecutionId,
		"level":        log.Level,
		"message":      log.Message,
		"timestamp":    log.CreatedAt,
	}

	result, err := g.DB().Model("workflow_execution_log").Ctx(ctx).Data(data).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	return result, nil
}

func (r *workflowRepository) ListLogs(ctx context.Context, workflowId int64, page, pageSize int) ([]*WorkflowLog, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	model := g.DB().Model("workflow_execution_log").Ctx(ctx).Where("workflow_id", workflowId).Safe()
	total, err := model.Count()
	if err != nil {
		return nil, 0, err
	}

	var records []struct {
		Id          int64  `json:"id"`
		WorkflowId  int64  `json:"workflow_id"`
		ExecutionId int64  `json:"execution_id"`
		Level       string `json:"level"`
		Message     string `json:"message"`
		Timestamp   int64  `json:"timestamp"`
	}
	err = model.Page(page, pageSize).Order("timestamp desc").Scan(&records)
	if err != nil {
		return nil, 0, err
	}
	logs := make([]*WorkflowLog, 0, len(records))
	for _, record := range records {
		logs = append(logs, &WorkflowLog{
			Id:          record.Id,
			WorkflowId:  record.WorkflowId,
			ExecutionId: record.ExecutionId,
			Level:       record.Level,
			Message:     record.Message,
			CreatedAt:   record.Timestamp,
		})
	}
	return logs, int(total), err
}

func executionStatusText(status int) string {
	switch status {
	case 1:
		return "running"
	case 2:
		return "completed"
	case 3:
		return "failed"
	default:
		return "pending"
	}
}

func executionStatusCode(status string) int {
	switch status {
	case "running":
		return 1
	case "completed":
		return 2
	case "failed":
		return 3
	default:
		return 0
	}
}
