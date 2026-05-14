package workflow

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (s *WorkflowService) DuplicateWorkflow(ctx context.Context, workflowId int64) (*Workflow, error) {
	workflow, err := WorkflowRepository().GetWorkflowById(ctx, workflowId)
	if err != nil {
		return nil, err
	}
	if workflow == nil {
		return nil, gerror.Newf("workflow %d not found", workflowId)
	}

	copyWorkflow := &Workflow{
		Name:        workflow.Name + " Copy",
		Description: workflow.Description,
		Status:      0,
		Version:     1,
		Trigger:     workflow.Trigger,
		Nodes:       workflow.Nodes,
		Connections: workflow.Connections,
		Metadata:    workflow.Metadata,
		CreatedAt:   time.Now().Unix(),
		UpdatedAt:   time.Now().Unix(),
	}

	newId, err := WorkflowRepository().CreateWorkflow(ctx, copyWorkflow)
	if err != nil {
		return nil, err
	}

	copyWorkflow.Id = newId
	return copyWorkflow, nil
}
