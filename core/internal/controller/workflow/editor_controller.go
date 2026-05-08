package workflow

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"

	v1 "github.com/BillionMail/BillionMail/core/api/workflow/v1"
	"github.com/BillionMail/BillionMail/core/internal/service"
)

type ControllerEditorV1 struct{}

func (c *ControllerEditorV1) Get(ctx context.Context, req *v1.GetWorkflowEditorReq) (*v1.GetWorkflowEditorRes, error) {
	result, err := service.WorkflowEditor().GetEditorData(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *ControllerEditorV1) Update(ctx context.Context, req *v1.UpdateWorkflowEditorReq) (*v1.UpdateWorkflowEditorRes, error) {
	err := service.WorkflowEditor().UpdateEditorData(ctx, req.Id, req.Nodes, req.Connections)
	if err != nil {
		return nil, gerror.Wrap(err, "failed to update workflow editor data")
	}
	return &v1.UpdateWorkflowEditorRes{}, nil
}