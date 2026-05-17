package workflow

import (
	workflowController "billionmail-core/internal/controller/workflow"

	"github.com/gogf/gf/v2/net/ghttp"
)

func RegisterRoutes(router *ghttp.RouterGroup) {
	router.Group("/workflow", func(group *ghttp.RouterGroup) {
		workflowV1Controller := workflowController.NewV1()
		workflowEditorController := workflowController.NewEditorV1()
		analyticsController := workflowController.NewAnalyticsController()
		webhookController := workflowController.NewWebhookController()

		group.Bind(workflowV1Controller)
		group.Bind(workflowEditorController)
		group.GET("/:id/editor", workflowEditorController.GetEditor)
		group.PUT("/:id/editor", workflowEditorController.UpdateEditor)
		group.POST("/:id/rollback/:version", workflowEditorController.Rollback)
		group.DELETE("/:id/versions/:version", workflowEditorController.DeleteVersion)
		group.POST("/:id/execution-log", workflowV1Controller.CreateExecutionLog)
		group.GET("/:id/execution-log", analyticsController.GetExecutionLog)
		group.GET("/:id/execution-log/export", workflowV1Controller.ExportExecutionLogs)
		group.GET("/:id/export", analyticsController.ExportExecutionLog)
		group.GET("/:id/node-stats", workflowV1Controller.GetNodeStats)
		group.GET("/:id/report", analyticsController.GetReport)
		group.POST("/trigger/:id", webhookController.Trigger)
	})
}
