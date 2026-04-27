package workflow

import (
	"context"
	"encoding/json"
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

	serialized, err := json.Marshal(map[string]interface{}{
		"name":        copyWorkflow.Name,
		"description": copyWorkflow.Description,
		"status":      copyWorkflow.Status,
		"trigger":     copyWorkflow.Trigger,
		"nodes":       copyWorkflow.Nodes,
		"connections": copyWorkflow.Connections,
		"metadata":    copyWorkflow.Metadata,
	})
	if err != nil {
		return nil, err
	}

	_, err = WorkflowRepository().CreateWorkflowVersion(ctx, &WorkflowVersion{
		WorkflowId:  newId,
		Version:     1,
		Name:        copyWorkflow.Name,
		Description: copyWorkflow.Description,
		Status:      copyWorkflow.Status,
		Trigger:     copyWorkflow.Trigger,
		Definition:  string(serialized),
		CreatedAt:   copyWorkflow.CreatedAt,
	})
	if err != nil {
		return nil, err
	}

	copyWorkflow.Id = newId
	return copyWorkflow, nil
}
