package workflow

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

type WorkflowService struct{}

var (
	workflowServiceInstance *WorkflowService
)

func GetWorkflowService() *WorkflowService {
	if workflowServiceInstance == nil {
		workflowServiceInstance = &WorkflowService{}
	}
	return workflowServiceInstance
}

func (s *WorkflowService) CreateWorkflow(ctx context.Context, workflow *Workflow) (int64, error) {
	if workflow == nil {
		return 0, gerror.New("workflow data is required")
	}
	if workflow.Name == "" {
		return 0, gerror.New("workflow name is required")
	}
	if workflow.Version <= 0 {
		workflow.Version = 1
	}
	workflow.CreatedAt = time.Now().Unix()
	workflow.UpdatedAt = workflow.CreatedAt

	workflowId, err := WorkflowRepository().CreateWorkflow(ctx, workflow)
	if err != nil {
		return 0, err
	}
	workflow.Id = workflowId

	serialized, err := json.Marshal(map[string]interface{}{
		"name":        workflow.Name,
		"description": workflow.Description,
		"status":      workflow.Status,
		"trigger":     workflow.Trigger,
		"nodes":       workflow.Nodes,
		"connections": workflow.Connections,
		"metadata":    workflow.Metadata,
	})
	if err != nil {
		return workflowId, err
	}

	_, err = WorkflowRepository().CreateWorkflowVersion(ctx, &WorkflowVersion{
		WorkflowId:  workflowId,
		Version:     workflow.Version,
		Name:        workflow.Name,
		Description: workflow.Description,
		Status:      workflow.Status,
		Trigger:     workflow.Trigger,
		Definition:  string(serialized),
		CreatedAt:   workflow.CreatedAt,
	})
	return workflowId, err
}

func (s *WorkflowService) UpdateWorkflow(ctx context.Context, workflow *Workflow) error {
	if workflow == nil {
		return gerror.New("workflow data is required")
	}
	if workflow.Id == 0 {
		return gerror.New("workflow id is required")
	}
	workflow.UpdatedAt = time.Now().Unix()

	current, err := WorkflowRepository().GetWorkflowById(ctx, workflow.Id)
	if err != nil {
		return err
	}
	if current == nil {
		return gerror.Newf("workflow %d not found", workflow.Id)
	}

	workflow.Version = current.Version + 1

	err = WorkflowRepository().UpdateWorkflow(ctx, workflow)
	if err != nil {
		return err
	}

	serialized, err := json.Marshal(map[string]interface{}{
		"name":        workflow.Name,
		"description": workflow.Description,
		"status":      workflow.Status,
		"trigger":     workflow.Trigger,
		"nodes":       workflow.Nodes,
		"connections": workflow.Connections,
		"metadata":    workflow.Metadata,
	})
	if err != nil {
		return err
	}

	_, err = WorkflowRepository().CreateWorkflowVersion(ctx, &WorkflowVersion{
		WorkflowId:  workflow.Id,
		Version:     workflow.Version,
		Name:        workflow.Name,
		Description: workflow.Description,
		Status:      workflow.Status,
		Trigger:     workflow.Trigger,
		Definition:  string(serialized),
		CreatedAt:   workflow.UpdatedAt,
	})
	return err
}

func (s *WorkflowService) GetWorkflow(ctx context.Context, workflowId int64) (*Workflow, error) {
	return WorkflowRepository().GetWorkflowById(ctx, workflowId)
}

func (s *WorkflowService) DeleteWorkflow(ctx context.Context, workflowId int64) error {
	return WorkflowRepository().DeleteWorkflow(ctx, workflowId)
}

func (s *WorkflowService) ListWorkflows(ctx context.Context, page, pageSize int, keyword string, status int) ([]*Workflow, int, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return WorkflowRepository().ListWorkflows(timeoutCtx, page, pageSize, keyword, status)
}

func (s *WorkflowService) ExecuteWorkflow(ctx context.Context, workflowId int64, trigger string) (*WorkflowExecution, error) {
	workflow, err := WorkflowRepository().GetWorkflowById(ctx, workflowId)
	if err != nil {
		return nil, err
	}
	if workflow == nil {
		return nil, gerror.Newf("workflow %d not found", workflowId)
	}

	execution := &WorkflowExecution{
		WorkflowId: workflowId,
		Version:    workflow.Version,
		Status:     1,
		Trigger:    trigger,
		StartedAt:  time.Now().Unix(),
		CreatedAt:  time.Now().Unix(),
		UpdatedAt:  time.Now().Unix(),
	}
	id, err := WorkflowRepository().RecordExecution(ctx, execution)
	if err != nil {
		return nil, err
	}
	execution.Id = id

	_, _ = WorkflowRepository().CreateLog(ctx, &WorkflowLog{
		WorkflowId:  workflowId,
		ExecutionId: execution.Id,
		Level:       "info",
		Message:     "Workflow execution started",
		CreatedAt:   time.Now().Unix(),
	})

	return execution, nil
}

func (s *WorkflowService) GetExecutionHistory(ctx context.Context, workflowId int64, page, pageSize int) ([]*WorkflowExecution, int, error) {
	return WorkflowRepository().ListExecutions(ctx, workflowId, page, pageSize)
}

func (s *WorkflowService) AddWorkflowLog(ctx context.Context, log *WorkflowLog) (int64, error) {
	if log == nil {
		return 0, gerror.New("workflow log is required")
	}
	log.CreatedAt = time.Now().Unix()
	return WorkflowRepository().CreateLog(ctx, log)
}

func (s *WorkflowService) GetWorkflowLogs(ctx context.Context, workflowId int64, page, pageSize int) ([]*WorkflowLog, int, error) {
	return WorkflowRepository().ListLogs(ctx, workflowId, page, pageSize)
}
