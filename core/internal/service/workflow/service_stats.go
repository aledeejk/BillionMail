package workflow

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

type WorkflowStatistics struct {
	WorkflowId      int64   `json:"workflow_id"`
	TotalExecutions int     `json:"total_executions"`
	SuccessCount    int     `json:"success_count"`
	FailureCount    int     `json:"failure_count"`
	AverageDuration float64 `json:"average_duration"`
	LastRunAt       int64   `json:"last_run_at"`
}

func (s *WorkflowService) GetWorkflowStatistics(ctx context.Context, workflowId int64) (*WorkflowStatistics, error) {
	var result struct {
		TotalExecutions int     `json:"total_executions"`
		SuccessCount    int     `json:"success_count"`
		FailureCount    int     `json:"failure_count"`
		AverageDuration float64 `json:"average_duration"`
		LastRunAt       int64   `json:"last_run_at"`
	}

	err := g.DB().Model("workflow_execution").Ctx(ctx).
		Fields(`COUNT(*) AS total_executions,
		SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) AS success_count,
		SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) AS failure_count,
		0 AS average_duration,
		COALESCE(MAX(started_at), 0) AS last_run_at`).
		Where("workflow_id", workflowId).
		Scan(&result)
	if err != nil {
		return nil, err
	}

	return &WorkflowStatistics{
		WorkflowId:      workflowId,
		TotalExecutions: result.TotalExecutions,
		SuccessCount:    result.SuccessCount,
		FailureCount:    result.FailureCount,
		AverageDuration: result.AverageDuration,
		LastRunAt:       result.LastRunAt,
	}, nil
}

func (s *WorkflowService) GetExecutionTimeline(ctx context.Context, workflowId int64, page, pageSize int) ([]*WorkflowExecution, int, error) {
	return WorkflowRepository().ListExecutions(ctx, workflowId, page, pageSize)
}

func (s *WorkflowService) GetWorkflowLogTimeline(ctx context.Context, workflowId int64, page, pageSize int) ([]*WorkflowLog, int, error) {
	return WorkflowRepository().ListLogs(ctx, workflowId, page, pageSize)
}
