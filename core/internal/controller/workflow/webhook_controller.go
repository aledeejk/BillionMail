package workflow

import (
	"encoding/json"

	workflowService "billionmail-core/internal/service/workflow"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"
)

type WebhookController struct{}

func NewWebhookController() *WebhookController {
	return &WebhookController{}
}

func (c *WebhookController) Trigger(r *ghttp.Request) {
	workflowId := gconv.Int64(r.Get("workflow_id"))
	if workflowId <= 0 {
		workflowId = gconv.Int64(r.Get("id"))
	}
	if workflowId <= 0 {
		r.Response.WriteJsonExit(g.Map{"code": 400, "msg": "workflow_id is required"})
		return
	}

	payload := map[string]interface{}{}
	if body := r.GetBodyString(); body != "" {
		if err := json.Unmarshal([]byte(body), &payload); err != nil {
			r.Response.WriteJsonExit(g.Map{"code": 400, "msg": err.Error()})
			return
		}
	}
	payload["trigger"] = "webhook"

	engine := workflowService.GetExecutionEngine()
	if engine.HasQueue() {
		if err := engine.DispatchExecution(r.Context(), workflowId, gconv.Int64(payload["contact_id"]), payload); err != nil {
			r.Response.WriteJsonExit(g.Map{"code": 500, "msg": err.Error()})
			return
		}
		r.Response.WriteJsonExit(g.Map{"status": "queued"})
		return
	}

	execution, err := engine.ExecuteWorkflow(r.Context(), workflowId, gconv.Int64(payload["contact_id"]), payload)
	if err != nil {
		r.Response.WriteJsonExit(g.Map{"code": 500, "msg": err.Error()})
		return
	}

	r.Response.WriteJsonExit(g.Map{"status": "started", "execution_id": gconv.String(execution.Id)})
}
