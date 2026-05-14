package workflow

import (
	v1 "billionmail-core/api/workflow/v1"
	workflowService "billionmail-core/internal/service/workflow"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type ExecutionController struct {
	service *workflowService.ExecutionService
}

func NewExecutionController() *ExecutionController {
	return &ExecutionController{service: workflowService.NewExecutionService()}
}

func (c *ExecutionController) GetExecutionLog(r *ghttp.Request) {
	var req v1.GetExecutionLogReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{"code": 400, "msg": err.Error()})
		return
	}

	list, total, err := c.service.GetExecutionLog(r.Context(), req.Id, req.Contact, req.Status)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{"code": 500, "msg": err.Error()})
		return
	}

	r.Response.WriteJsonExit(v1.GetExecutionLogRes{List: toV1ExecutionLogItems(list), Total: total})
}

func (c *ExecutionController) GetReport(r *ghttp.Request) {
	var req v1.GetReportReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{"code": 400, "msg": err.Error()})
		return
	}

	data, err := c.service.GetReport(r.Context(), req.Id)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{"code": 500, "msg": err.Error()})
		return
	}

	r.Response.WriteJsonExit(v1.GetReportRes{
		Sent:         data.Sent,
		UniqueOpens:  data.UniqueOpens,
		UniqueClicks: data.UniqueClicks,
		Unsubscribes: data.Unsubscribes,
	})
}

func toV1ExecutionLogItems(items []workflowService.ExecutionLogItem) []v1.ExecutionLogItem {
	result := make([]v1.ExecutionLogItem, 0, len(items))
	for _, item := range items {
		nodes := make([]v1.LogNode, 0, len(item.Nodes))
		for _, node := range item.Nodes {
			nodes = append(nodes, v1.LogNode{
				NodeId:    node.NodeId,
				NodeType:  node.NodeType,
				Status:    node.Status,
				Timestamp: node.Timestamp,
			})
		}
		finishedAt := item.FinishedAt
		result = append(result, v1.ExecutionLogItem{
			ExecutionId:  item.ExecutionId,
			ContactId:    item.ContactId,
			ContactEmail: item.ContactEmail,
			Status:       item.Status,
			StartedAt:    item.StartedAt,
			FinishedAt:   &finishedAt,
			Nodes:        nodes,
		})
	}
	return result
}
