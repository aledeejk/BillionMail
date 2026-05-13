package workflow

import (
	workflowController "billionmail-core/internal/controller/workflow"

	"github.com/gogf/gf/v2/net/ghttp"
)

func RegisterRoutes(router *ghttp.RouterGroup) {
	router.Group("/workflow", func(group *ghttp.RouterGroup) {
		workflowEditorController := workflowController.NewEditorV1()

		// Workflow endpoints
		// POST   /api/workflow
		// GET    /api/workflow
		// GET    /api/workflow/{id}
		// PUT    /api/workflow/{id}
		// DELETE /api/workflow/{id}
		// POST   /api/workflow/{id}/duplicate
		// POST   /api/workflow/{id}/toggle
		// POST   /api/workflow/{id}/execute
		// GET    /api/workflow/{id}/stats
		// GET    /api/workflow/{id}/versions
		// GET    /api/workflow/{id}/executions
		group.Bind(workflowController.NewV1())
		group.Bind(workflowEditorController)
		group.GET("/:id/editor", workflowEditorController.GetEditor)
		group.PUT("/:id/editor", workflowEditorController.UpdateEditor)
	})
}
