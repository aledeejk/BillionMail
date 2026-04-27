package workflow

import (
	workflowController "billionmail-core/internal/controller/workflow"

	"github.com/gogf/gf/v2/net/ghttp"
)

func RegisterRoutes(router *ghttp.RouterGroup) {
	router.Group("/workflow", func(group *ghttp.RouterGroup) {
		// Workflow endpoints
		// POST /api/workflow
		// GET /api/workflow
		// GET /api/workflow/{id}
		// PUT /api/workflow/{id}
		// DELETE /api/workflow/{id}
		// POST /api/workflow/{id}/duplicate
		// POST /api/workflow/{id}/toggle
		// GET /api/workflow/{id}/stats
		group.Bind(workflowController.NewV1())
	})
}
