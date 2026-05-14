package workflow

import (
	"fmt"

	v1 "billionmail-core/api/workflow/v1"
	workflowService "billionmail-core/internal/service/workflow"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"
)

type AnalyticsController struct{}

func NewAnalyticsController() *AnalyticsController {
	return &AnalyticsController{}
}

func (c *AnalyticsController) GetExecutionLog(r *ghttp.Request) {
	fmt.Println("GetExecutionLog called")
	workflowId := gconv.Int64(r.Get("id"))
	items, err := workflowService.GetWorkflowService().GetExecutionWalkthroughs(r.Context(), workflowId, workflowService.WorkflowLogFilter{
		Contact: r.Get("contact").String(),
		Status:  r.Get("status").String(),
	})
	if err != nil {
		r.Response.WriteJsonExit(g.Map{"code": 500, "msg": err.Error()})
		return
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

	r.Response.WriteJsonExit(g.Map{"list": res, "total": len(res)})
}

func (c *AnalyticsController) GetReport(r *ghttp.Request) {
	fmt.Println("GetReport called")
	workflowId := gconv.Int64(r.Get("id"))
	report, err := workflowService.GetWorkflowService().GetWorkflowReport(r.Context(), workflowId)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{"code": 500, "msg": err.Error()})
		return
	}

	r.Response.WriteJsonExit(g.Map{
		"sent":          report.EmailsSent,
		"emails_sent":   report.EmailsSent,
		"unique_opens":  report.UniqueOpens,
		"unique_clicks": report.UniqueClicks,
		"unsubscribes":  report.Unsubscribes,
		"tracking_stub": report.TrackingStub,
	})
}

func (c *AnalyticsController) ExportExecutionLog(r *ghttp.Request) {
	var req v1.ExportExecutionLogReq
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJsonExit(g.Map{"code": 400, "msg": err.Error()})
		return
	}
	if req.Id == 0 {
		req.Id = gconv.Int64(r.Get("id"))
	}

	csvData, err := workflowService.NewAnalyticsService().ExportExecutionLogCSV(r.Context(), req.Id, req.Contact, req.Status)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{"code": 500, "msg": err.Error()})
		return
	}

	r.Response.Header().Set("Content-Type", "text/csv; charset=utf-8")
	r.Response.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=execution_log_%d.csv", req.Id))
	r.Response.Write(csvData)
}
