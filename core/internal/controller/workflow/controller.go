package workflow

import (
	"context"
	"errors"
	"fmt"
	"time"

	v1 "billionmail-core/api/workflow/v1"
	workflowService "billionmail-core/internal/service/workflow"

	"github.com/gogf/gf/v2/frame/g"
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
	if err := workflowService.GetWorkflowService().ToggleWorkflow(ctx, workflowId); err != nil {
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

func (c *ControllerV1) GetVersions(ctx context.Context, req *v1.GetWorkflowVersionsReq) ([]*v1.WorkflowVersionRes, error) {
	workflowId := gconv.Int64(req.Id)
	versions, err := workflowService.GetWorkflowService().GetWorkflowVersions(ctx, workflowId)
	if err != nil {
		return nil, err
	}
	res := make([]*v1.WorkflowVersionRes, 0, len(versions))
	for _, version := range versions {
		res = append(res, toV1WorkflowVersion(version))
	}
	return res, nil
}

func (c *ControllerV1) GetExecutions(ctx context.Context, req *v1.GetWorkflowExecutionsReq) ([]*v1.WorkflowExecutionRes, error) {
	workflowId := gconv.Int64(req.Id)
	executions, _, err := workflowService.GetWorkflowService().GetExecutionHistory(ctx, workflowId, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}
	res := make([]*v1.WorkflowExecutionRes, 0, len(executions))
	for _, execution := range executions {
		res = append(res, toV1WorkflowExecution(execution))
	}
	return res, nil
}

func (c *ControllerV1) Execute(ctx context.Context, req *v1.ExecuteWorkflowReq) (*v1.WorkflowExecutionRes, error) {
	workflowId := gconv.Int64(req.Id)
	inputData := req.InputData
	if inputData == nil {
		inputData = map[string]interface{}{}
	}
	if req.Trigger != "" {
		inputData["trigger"] = req.Trigger
	}
	contactEmail := req.ContactEmail
	if contactEmail == "" {
		contactEmail = gconv.String(inputData["contact_email"])
		if contactEmail == "" {
			contactEmail = gconv.String(inputData["email"])
		}
	}
	if contactEmail != "" {
		inputData["contact_email"] = contactEmail
		inputData["email"] = contactEmail
	}
	if gconv.String(inputData["idempotency_key"]) == "" {
		minuteWindow := time.Now().Unix() / 60
		inputData["idempotency_key"] = fmt.Sprintf("%d:%s:%d", workflowId, contactEmail, minuteWindow)
	}
	engine := workflowService.GetExecutionEngine()
	if engine.HasQueue() {
		if dispatchErr := engine.DispatchExecution(ctx, workflowId, req.ContactId, inputData); dispatchErr != nil {
			return nil, dispatchErr
		}
		return &v1.WorkflowExecutionRes{Status: 0}, nil
	}
	execution, err := engine.ExecuteWorkflow(ctx, workflowId, req.ContactId, inputData)
	if err != nil {
		return nil, err
	}
	return toV1WorkflowExecution(execution), nil
}

func (c *ControllerV1) CreateExecutionLog(ctx context.Context, req *v1.CreateWorkflowExecutionLogReq) (*v1.WorkflowExecutionLogRes, error) {
	workflowId := gconv.Int64(req.Id)
	log := &workflowService.WorkflowLog{
		WorkflowId:   workflowId,
		ExecutionId:  req.ExecutionId,
		ContactId:    req.ContactId,
		NodeId:       req.NodeId,
		NodeType:     req.NodeType,
		Status:       req.Status,
		Message:      req.Message,
		StartedAt:    req.StartedAt,
		FinishedAt:   req.FinishedAt,
		ErrorMessage: req.ErrorMessage,
	}
	id, err := workflowService.GetWorkflowService().CreateExecutionLog(ctx, log)
	if err != nil {
		return nil, err
	}
	log.Id = id
	return toV1WorkflowLog(log), nil
}

func (c *ControllerV1) GetExecutionLogs(ctx context.Context, req *v1.GetWorkflowExecutionLogsReq) (*v1.WorkflowExecutionLogListRes, error) {
	workflowId := gconv.Int64(req.Id)
	items, err := workflowService.GetWorkflowService().GetExecutionWalkthroughs(ctx, workflowId, workflowService.WorkflowLogFilter{
		Contact: req.Contact,
		Status:  req.Status,
		From:    req.From,
		To:      req.To,
	})
	if err != nil {
		return nil, err
	}
	res := make([]*v1.WorkflowExecutionWalkthroughRes, 0, len(items))
	for _, item := range items {
		nodes := make([]*v1.WorkflowExecutionNodeLogRes, 0, len(item.Nodes))
		for _, node := range item.Nodes {
			nodes = append(nodes, &v1.WorkflowExecutionNodeLogRes{
				NodeId:       node.NodeId,
				NodeType:     node.NodeType,
				Status:       node.Status,
				Timestamp:    node.Timestamp,
				ErrorMessage: node.ErrorMessage,
			})
		}
		res = append(res, &v1.WorkflowExecutionWalkthroughRes{
			ExecutionId:  item.ExecutionId,
			ContactId:    item.ContactId,
			ContactEmail: item.ContactEmail,
			Status:       item.Status,
			StartedAt:    item.StartedAt,
			FinishedAt:   item.FinishedAt,
			Nodes:        nodes,
		})
	}
	return &v1.WorkflowExecutionLogListRes{List: res, Total: len(res)}, nil
}

func (c *ControllerV1) GetNodeStats(ctx context.Context, req *v1.GetWorkflowNodeStatsReq) ([]workflowService.WorkflowNodeStatistic, error) {
	return workflowService.GetWorkflowService().GetNodeStatistics(ctx, gconv.Int64(req.Id))
}

func (c *ControllerV1) GetReport(ctx context.Context, req *v1.GetWorkflowReportReq) (*v1.WorkflowReportRes, error) {
	report, err := workflowService.GetWorkflowService().GetWorkflowReport(ctx, gconv.Int64(req.Id))
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowReportRes{
		Sent:         report.EmailsSent,
		EmailsSent:   report.EmailsSent,
		UniqueOpens:  report.UniqueOpens,
		UniqueClicks: report.UniqueClicks,
		Unsubscribes: report.Unsubscribes,
		TrackingStub: report.TrackingStub,
	}, nil
}

func (c *ControllerV1) ExportExecutionLogs(ctx context.Context, req *v1.ExportWorkflowExecutionLogsReq) (*struct{}, error) {
	workflowId := gconv.Int64(req.Id)
	content, err := workflowService.GetWorkflowService().ExportExecutionLogsCSV(ctx, workflowId, workflowService.WorkflowLogFilter{
		Contact: req.Contact,
		Status:  req.Status,
		From:    req.From,
		To:      req.To,
	})
	if err != nil {
		return nil, err
	}
	response := g.RequestFromCtx(ctx).Response
	response.Header().Set("Content-Type", "text/csv")
	response.Header().Set("Content-Disposition", "attachment; filename="+workflowService.FormatWorkflowCSVFilename(workflowId))
	response.Write(content)
	return nil, nil
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

func toV1WorkflowVersion(version *workflowService.WorkflowVersion) *v1.WorkflowVersionRes {
	if version == nil {
		return nil
	}
	return &v1.WorkflowVersionRes{
		Version:   version.Version,
		CreatedAt: time.Unix(version.CreatedAt, 0),
		Status:    version.Status,
	}
}

func toV1WorkflowExecution(execution *workflowService.WorkflowExecution) *v1.WorkflowExecutionRes {
	if execution == nil {
		return nil
	}
	return &v1.WorkflowExecutionRes{
		Id:           gconv.String(execution.Id),
		WorkflowId:   gconv.String(execution.WorkflowId),
		Version:      execution.Version,
		Status:       execution.Status,
		Trigger:      execution.Trigger,
		StartedAt:    time.Unix(execution.StartedAt, 0),
		CompletedAt:  time.Unix(execution.CompletedAt, 0),
		ErrorMessage: execution.Error,
	}
}

func toV1WorkflowLog(log *workflowService.WorkflowLog) *v1.WorkflowExecutionLogRes {
	if log == nil {
		return nil
	}
	return &v1.WorkflowExecutionLogRes{
		Id:           gconv.String(log.Id),
		WorkflowId:   gconv.String(log.WorkflowId),
		ExecutionId:  log.ExecutionId,
		ContactId:    log.ContactId,
		NodeId:       log.NodeId,
		NodeType:     log.NodeType,
		Status:       log.Status,
		Message:      log.Message,
		StartedAt:    log.StartedAt,
		FinishedAt:   log.FinishedAt,
		ErrorMessage: log.ErrorMessage,
	}
}

func boolToStatus(active bool) int {
	if active {
		return 1
	}
	return 0
}
