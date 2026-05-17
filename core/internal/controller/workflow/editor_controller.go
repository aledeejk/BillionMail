package workflow

import (
	"context"
	"errors"
	"net/http"

	v1 "billionmail-core/api/workflow/v1"
	workflowService "billionmail-core/internal/service/workflow"
	"billionmail-core/utility/types/api_v1"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
)

type ControllerEditorV1 struct{}

func NewEditorV1() *ControllerEditorV1 {
	return &ControllerEditorV1{}
}

func (c *ControllerEditorV1) GetEditor(r *ghttp.Request) {
	result, err := workflowService.WorkflowEditor().GetEditorData(r.Context(), r.Get("id").String())
	if err != nil {
		r.Response.WriteJsonExit(api_v1.StandardRes{
			Success: false,
			Code:    500,
			Msg:     err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(api_v1.StandardRes{
		Success: true,
		Code:    0,
		Msg:     "Success",
		Data:    result,
	})
}

func (c *ControllerEditorV1) UpdateEditor(r *ghttp.Request) {
	req := &v1.UpdateWorkflowEditorReq{
		Id: r.Get("id").String(),
	}

	if err := r.Parse(req); err != nil {
		r.Response.WriteJsonExit(api_v1.StandardRes{
			Success: false,
			Code:    400,
			Msg:     err.Error(),
		})
		return
	}

	if err := workflowService.WorkflowEditor().UpdateEditorData(r.Context(), req.Id, req.Nodes, req.Connections); err != nil {
		r.Response.WriteJsonExit(api_v1.StandardRes{
			Success: false,
			Code:    500,
			Msg:     err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(api_v1.StandardRes{
		Success: true,
		Code:    0,
		Msg:     "Workflow editor saved successfully",
		Data:    &v1.UpdateWorkflowEditorRes{},
	})
}

func (c *ControllerEditorV1) Rollback(r *ghttp.Request) {
	id := r.Get("id").String()
	version := r.Get("version").Int()
	if id == "" || version <= 0 {
		r.Response.WriteJsonExit(api_v1.StandardRes{
			Success: false,
			Code:    400,
			Msg:     "Invalid parameters",
		})
		return
	}

	editorData, err := workflowService.WorkflowEditor().RollbackEditorData(r.Context(), id, version)
	if err != nil {
		r.Response.WriteJsonExit(api_v1.StandardRes{
			Success: false,
			Code:    500,
			Msg:     err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(api_v1.StandardRes{
		Success: true,
		Code:    0,
		Msg:     "Workflow rolled back successfully",
		Data:    editorData,
	})
}

func (c *ControllerEditorV1) DeleteVersion(r *ghttp.Request) {
	err := workflowService.WorkflowEditor().DeleteWorkflowVersion(r.Context(), r.Get("id").String(), r.Get("version").Int())
	if err != nil {
		code := http.StatusInternalServerError
		if errors.Is(err, workflowService.ErrCurrentWorkflowVersion) {
			code = http.StatusBadRequest
		}
		if errors.Is(err, workflowService.ErrWorkflowVersionNotFound) {
			code = http.StatusNotFound
		}
		r.Response.Status = code
		r.Response.WriteJsonExit(api_v1.StandardRes{
			Success: false,
			Code:    code,
			Msg:     err.Error(),
		})
		return
	}

	r.Response.WriteJsonExit(api_v1.StandardRes{
		Success: true,
		Code:    0,
		Msg:     "Version deleted",
	})
}

func (c *ControllerEditorV1) Get(ctx context.Context, req *v1.GetWorkflowEditorReq) (*api_v1.StandardRes, error) {
	result, err := workflowService.WorkflowEditor().GetEditorData(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &api_v1.StandardRes{
		Success: true,
		Code:    0,
		Msg:     "Success",
		Data:    result,
	}, nil
}

func (c *ControllerEditorV1) Update(ctx context.Context, req *v1.UpdateWorkflowEditorReq) (*api_v1.StandardRes, error) {
	err := workflowService.WorkflowEditor().UpdateEditorData(ctx, req.Id, req.Nodes, req.Connections)
	if err != nil {
		return nil, gerror.Wrap(err, "failed to update workflow editor data")
	}
	return &api_v1.StandardRes{
		Success: true,
		Code:    0,
		Msg:     "Workflow editor saved successfully",
		Data:    &v1.UpdateWorkflowEditorRes{},
	}, nil
}
