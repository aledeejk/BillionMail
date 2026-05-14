package workflow

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type ExecutionLogItem struct {
	ExecutionId  string    `json:"execution_id"`
	ContactId    string    `json:"contact_id"`
	ContactEmail string    `json:"contact_email"`
	Status       string    `json:"status"`
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   time.Time `json:"finished_at"`
	Nodes        []LogNode `json:"nodes"`
}

type LogNode struct {
	NodeId    string    `json:"node_id"`
	NodeType  string    `json:"node_type"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type ReportData struct {
	Sent         int
	UniqueOpens  int
	UniqueClicks int
	Unsubscribes int
}

type ExecutionService struct{}

func NewExecutionService() *ExecutionService {
	return &ExecutionService{}
}

func (s *ExecutionService) GetExecutionLog(ctx context.Context, workflowId int64, contact, status string) ([]ExecutionLogItem, int, error) {
	if err := GetWorkflowService().EnsureExecutionLogTable(ctx); err != nil {
		return nil, 0, err
	}
	if !tableExists(ctx, "workflow_execution_log") {
		return []ExecutionLogItem{}, 0, nil
	}

	model := g.DB().Model("workflow_execution_log l").Ctx(ctx).
		LeftJoin("workflow_execution e", "e.id = l.execution_id").
		Where("l.workflow_id", workflowId).
		Safe()
	if contact != "" {
		model = model.Where("(l.contact_id LIKE ? OR e.contact_email LIKE ?)", "%"+contact+"%", "%"+contact+"%")
	}
	if status != "" {
		model = model.Where("l.status", status)
	}

	var rows []struct {
		WorkflowLog
		ExecutionContactEmail string `json:"execution_contact_email"`
		ExecutionContactId    string `json:"execution_contact_id"`
		LoggedContactEmail    string `json:"logged_contact_email"`
	}
	err := model.Fields("l.id, l.workflow_id, l.execution_id, l.contact_id, l.node_id, l.node_type, l.status, l.started_at, l.finished_at, l.error_message, l.timestamp AS created_at, e.contact_email AS execution_contact_email, e.contact_id_text AS execution_contact_id, (SELECT wl.details::jsonb->>'to' FROM workflow_log wl WHERE wl.execution_id = l.execution_id::text AND wl.action = 'email_sent' AND wl.details <> '' LIMIT 1) AS logged_contact_email").Order("l.started_at desc, l.id desc").Scan(&rows)
	if err != nil {
		return nil, 0, err
	}
	if len(rows) == 0 {
		return []ExecutionLogItem{}, 0, nil
	}

	items := make([]ExecutionLogItem, 0)
	index := map[string]int{}
	for _, row := range rows {
		key := executionKey(&row.WorkflowLog)
		idx, ok := index[key]
		startedAt := unixTime(row.StartedAt)
		finishedAt := unixTime(row.FinishedAt)
		contactEmail := row.ExecutionContactEmail
		if contactEmail == "" {
			contactEmail = row.LoggedContactEmail
		}
		if contactEmail == "" || contactEmail == "0" {
			contactEmail = row.ContactId
		}
		contactId := row.ExecutionContactId
		if contactId == "" {
			contactId = row.ContactId
		}
		if !ok {
			index[key] = len(items)
			items = append(items, ExecutionLogItem{
				ExecutionId:  key,
				ContactId:    contactId,
				ContactEmail: contactEmail,
				Status:       normalizeExecutionStatus(row.Status),
				StartedAt:    startedAt,
				FinishedAt:   finishedAt,
				Nodes:        []LogNode{},
			})
			idx = len(items) - 1
		}
		if row.Status == "failed" {
			items[idx].Status = "failed"
		}
		if items[idx].StartedAt.IsZero() || (!startedAt.IsZero() && startedAt.Before(items[idx].StartedAt)) {
			items[idx].StartedAt = startedAt
		}
		if finishedAt.After(items[idx].FinishedAt) {
			items[idx].FinishedAt = finishedAt
		}
		items[idx].Nodes = append(items[idx].Nodes, LogNode{
			NodeId:    row.NodeId,
			NodeType:  row.NodeType,
			Status:    row.Status,
			Timestamp: startedAt,
		})
	}

	return items, len(items), nil
}

func (s *ExecutionService) GetReport(ctx context.Context, workflowId int64) (*ReportData, error) {
	report := &ReportData{}
	if tableExists(ctx, "workflow_execution_log") {
		count, err := g.DB().Model("workflow_execution_log").Ctx(ctx).
			Where("workflow_id", workflowId).
			WhereIn("node_type", []string{"send-email", "email"}).
			Where("status", "success").
			Count()
		if err != nil {
			return nil, err
		}
		report.Sent = count
	}
	if tableExists(ctx, "mail_open") {
		count, _ := g.DB().Model("mail_open").Ctx(ctx).Fields("COUNT(DISTINCT contact_id)").Count()
		report.UniqueOpens = count
	}
	if tableExists(ctx, "mail_click") {
		count, _ := g.DB().Model("mail_click").Ctx(ctx).Fields("COUNT(DISTINCT contact_id)").Count()
		report.UniqueClicks = count
	}
	if tableExists(ctx, "unsubscribe_records") {
		count, _ := g.DB().Model("unsubscribe_records").Ctx(ctx).Count()
		report.Unsubscribes = count
	}
	return report, nil
}

func executionKey(row *WorkflowLog) string {
	if row.ExecutionId > 0 {
		return strconv.FormatInt(row.ExecutionId, 10)
	}
	if row.ContactId != "" {
		return fmt.Sprintf("%d-%s", row.WorkflowId, row.ContactId)
	}
	return strconv.FormatInt(row.Id, 10)
}

func unixTime(value int64) time.Time {
	if value <= 0 {
		return time.Time{}
	}
	return time.Unix(value, 0)
}

func normalizeExecutionStatus(status string) string {
	if status == "success" {
		return "completed"
	}
	if status == "" {
		return "running"
	}
	return status
}
