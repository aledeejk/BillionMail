package workflow

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
)

func (s *WorkflowService) GetWorkflowVersions(ctx context.Context, workflowId int64) ([]*WorkflowVersion, error) {
	return WorkflowRepository().GetWorkflowVersions(ctx, workflowId)
}

func (s *WorkflowService) RollbackWorkflow(ctx context.Context, workflowId int64, versionId int64) error {
	version, err := WorkflowRepository().GetWorkflowVersionById(ctx, versionId)
	if err != nil {
		return err
	}
	if version == nil || version.WorkflowId != workflowId {
		return gerror.Newf("workflow version %d not found for workflow %d", versionId, workflowId)
	}

	var definition struct {
		Name        string                 `json:"name"`
		Description string                 `json:"description"`
		Status      int                    `json:"status"`
		Trigger     string                 `json:"trigger"`
		Nodes       []*WorkflowNode        `json:"nodes"`
		Connections []*WorkflowConnection  `json:"connections"`
		Metadata    map[string]interface{} `json:"metadata"`
	}
	if err := json.Unmarshal([]byte(version.Definition), &definition); err != nil {
		return err
	}

	workflow := &Workflow{
		Id:          workflowId,
		Name:        definition.Name,
		Description: definition.Description,
		Status:      definition.Status,
		Trigger:     definition.Trigger,
		Nodes:       definition.Nodes,
		Connections: definition.Connections,
		Metadata:    definition.Metadata,
		Version:     version.Version,
		UpdatedAt:   time.Now().Unix(),
	}

	if err := WorkflowRepository().UpdateWorkflow(ctx, workflow); err != nil {
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
		WorkflowId:  workflowId,
		Version:     workflow.Version,
		Name:        workflow.Name,
		Description: workflow.Description,
		Status:      workflow.Status,
		Trigger:     workflow.Trigger,
		Definition:  string(serialized),
		CreatedAt:   time.Now().Unix(),
	})
	return err
}

func (s *WorkflowService) CreateWorkflowVersion(ctx context.Context, version *WorkflowVersion) (int64, error) {
	if version == nil {
		return 0, gerror.New("workflow version is required")
	}
	if version.WorkflowId == 0 {
		return 0, gerror.New("workflow id is required")
	}
	if version.Version <= 0 {
		return 0, gerror.New("version number is required")
	}
	version.CreatedAt = time.Now().Unix()
	return WorkflowRepository().CreateWorkflowVersion(ctx, version)
}
