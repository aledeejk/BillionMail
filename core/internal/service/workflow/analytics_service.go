package workflow

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

type NodeStat struct {
	NodeId         string  `json:"node_id"`
	NodeType       string  `json:"node_type"`
	EnteredCount   int     `json:"entered_count"`
	CompletedCount int     `json:"completed_count"`
	ConversionRate float64 `json:"conversion_rate"`
	AvgDurationMs  int64   `json:"avg_duration_ms"`
}

type AnalyticsService struct{}

func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{}
}

func (s *AnalyticsService) GetExecutionLog(ctx context.Context, workflowId int64, contact string, status string) ([]ExecutionLogItem, int, error) {
	service := NewExecutionService()
	return service.GetExecutionLog(ctx, workflowId, contact, status)
}

func (s *AnalyticsService) GetReport(ctx context.Context, workflowId int64) (*ReportData, error) {
	service := NewExecutionService()
	return service.GetReport(ctx, workflowId)
}

func (s *AnalyticsService) GetNodeStats(ctx context.Context, workflowId int64) ([]NodeStat, error) {
	if !tableExists(ctx, "workflow_execution_log") {
		return []NodeStat{}, nil
	}

	var rows []*WorkflowLog
	err := g.DB().Model("workflow_execution_log").Ctx(ctx).
		Where("workflow_id", workflowId).
		WhereNot("node_id", "").
		Fields("node_id, node_type, status, started_at, finished_at").
		Order("node_id asc, started_at asc").
		Scan(&rows)
	if err != nil {
		return nil, err
	}

	stats := make([]NodeStat, 0)
	index := map[string]int{}
	for _, row := range rows {
		idx, ok := index[row.NodeId]
		if !ok {
			index[row.NodeId] = len(stats)
			stats = append(stats, NodeStat{NodeId: row.NodeId, NodeType: row.NodeType})
			idx = len(stats) - 1
		}
		stats[idx].EnteredCount++
		if row.Status == "success" || row.Status == "completed" {
			stats[idx].CompletedCount++
		}
		if row.FinishedAt > row.StartedAt {
			stats[idx].AvgDurationMs += (row.FinishedAt - row.StartedAt) * 1000
		}
	}

	for i := range stats {
		if stats[i].EnteredCount > 0 {
			stats[i].ConversionRate = float64(stats[i].CompletedCount) / float64(stats[i].EnteredCount) * 100
			stats[i].AvgDurationMs = stats[i].AvgDurationMs / int64(stats[i].EnteredCount)
		}
	}
	return stats, nil
}

func (s *AnalyticsService) ExportExecutionLogCSV(ctx context.Context, workflowId int64, contact, status string) ([]byte, error) {
	items, err := GetWorkflowService().GetExecutionWalkthroughs(ctx, workflowId, WorkflowLogFilter{
		Contact: contact,
		Status:  status,
	})
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	if err := writer.Write([]string{"Execution ID", "Contact Email", "Status", "Started At", "Finished At", "Nodes"}); err != nil {
		return nil, err
	}
	for _, item := range items {
		nodes := make([]string, 0, len(item.Nodes))
		for _, node := range item.Nodes {
			nodes = append(nodes, fmt.Sprintf("%s (%s) %s", node.NodeType, node.NodeId, node.Status))
		}
		contactEmail := item.ContactEmail
		if contactEmail == "" {
			contactEmail = item.ContactId
		}
		if err := writer.Write([]string{
			item.ExecutionId,
			contactEmail,
			item.Status,
			item.StartedAt,
			item.FinishedAt,
			strings.Join(nodes, "; "),
		}); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
