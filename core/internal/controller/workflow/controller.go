package workflow

import (
	"context"
	"errors"
	"time"

	v1 "billionmail-core/api/workflow/v1"
	workflowService "billionmail-core/internal/service/workflow"
	"github.com/gogf/gf/v2/util/gconv"
)

type ControllerV1 struct{}

func NewV1() *ControllerV1 {
	return &ControllerV1{}
}

func (c *ControllerV1) Create(ctx context.Context, req *v1.CreateWorkflowReq) (*v1.WorkflowRes, error) {
	workflowEntity := &workflowService.Workflow{
		Name:        req.Name,
		Description: req.Description,
		Status:      boolToStatus(req.IsActive),
	}

	id, err := workflowService.GetWorkflowService().CreateWorkflow(ctx, workflowEntity)
	if err != nil {
		return nil, err
	}

	createdWorkflow, err := workflowService.GetWorkflowService().GetWorkflow(ctx, id)
	if err != nil {
		return nil, err
	}
	if createdWorkflow == nil {
		return nil, errors.New("workflow not found after create")
	}

	return toV1Workflow(createdWorkflow), nil
}

func (c *ControllerV1) Update(ctx context.Context, req *v1.UpdateWorkflowReq) (*v1.WorkflowRes, error) {
	workflowId := gconv.Int64(req.Id)
	workflowEntity := &workflowService.Workflow{
		Id:          workflowId,
		Name:        req.Name,
		Description: req.Description,
		Status:      boolToStatus(req.IsActive),
	}

	if err := workflowService.GetWorkflowService().UpdateWorkflow(ctx, workflowEntity); err != nil {
		return nil, err
	}

	updatedWorkflow, err := workflowService.GetWorkflowService().GetWorkflow(ctx, workflowId)
	if err != nil {
		return nil, err
	}
	if updatedWorkflow == nil {
		return nil, errors.New("workflow not found after update")
	}

	return toV1Workflow(updatedWorkflow), nil
}

func (c *ControllerV1) Get(ctx context.Context, req *v1.GetWorkflowReq) (*v1.WorkflowRes, error) {
	workflowId := gconv.Int64(req.Id)
	workflowEntity, err := workflowService.GetWorkflowService().GetWorkflow(ctx, workflowId)
	if err != nil {
		return nil, err
	}
	if workflowEntity == nil {
		return nil, errors.New("workflow not found")
	}
	return toV1Workflow(workflowEntity), nil
}

func (c *ControllerV1) GetList(ctx context.Context, req *v1.GetWorkflowListReq) (*v1.WorkflowListRes, error) {
	workflows, total, err := workflowService.GetWorkflowService().ListWorkflows(ctx, req.Page, req.Limit, req.Search, -1)
	if err != nil {
		return nil, err
	}

	res := &v1.WorkflowListRes{
		List:  make([]v1.WorkflowRes, 0, len(workflows)),
		Total: total,
	}
	for _, item := range workflows {
		res.List = append(res.List, *toV1Workflow(item))
	}
	return res, nil
}

func (c *ControllerV1) Delete(ctx context.Context, req *v1.DeleteWorkflowReq) (*struct{}, error) {
	workflowId := gconv.Int64(req.Id)
	if err := workflowService.GetWorkflowService().DeleteWorkflow(ctx, workflowId); err != nil {
		return nil, err
	}
	return &struct{}{}, nil
}

func (c *ControllerV1) Duplicate(ctx context.Context, req *v1.DuplicateWorkflowReq) (*v1.DuplicateWorkflowRes, error) {
	workflowId := gconv.Int64(req.Id)
	workflowEntity, err := workflowService.GetWorkflowService().DuplicateWorkflow(ctx, workflowId)
	if err != nil {
		return nil, err
	}
	return &v1.DuplicateWorkflowRes{Workflow: *toV1Workflow(workflowEntity)}, nil
}

func (c *ControllerV1) Toggle(ctx context.Context, req *v1.ToggleWorkflowReq) (*struct{}, error) {
	workflowId := gconv.Int64(req.Id)
	workflowEntity, err := workflowService.GetWorkflowService().GetWorkflow(ctx, workflowId)
	if err != nil {
		return nil, err
	}
	if workflowEntity == nil {
		return nil, errors.New("workflow not found")
	}

	workflowEntity.Status = 1 - workflowEntity.Status
	if err := workflowService.GetWorkflowService().UpdateWorkflow(ctx, workflowEntity); err != nil {
		return nil, err
	}

	return &struct{}{}, nil
}

func (c *ControllerV1) GetStats(ctx context.Context, req *v1.GetWorkflowStatsReq) (*v1.WorkflowStatsRes, error) {
	workflowId := gconv.Int64(req.Id)
	stats, err := workflowService.GetWorkflowService().GetWorkflowStatistics(ctx, workflowId)
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowStatsRes{
		ActiveContacts: stats.TotalExecutions,
		EmailsSent:     stats.SuccessCount,
		ConversionRate: 0,
	}, nil
}

func toV1Workflow(workflowEntity *workflowService.Workflow) *v1.WorkflowRes {
	if workflowEntity == nil {
		return nil
	}
	return &v1.WorkflowRes{
		Id:          gconv.String(workflowEntity.Id),
		Name:        workflowEntity.Name,
		Description: workflowEntity.Description,
		IsActive:    workflowEntity.Status == 1,
		Version:     workflowEntity.Version,
		CreatedAt:   time.Unix(workflowEntity.CreatedAt, 0),
		UpdatedAt:   time.Unix(workflowEntity.UpdatedAt, 0),
	}
}

func boolToStatus(active bool) int {
	if active {
		return 1
	}
	return 0
}
