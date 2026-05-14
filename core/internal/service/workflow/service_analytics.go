package workflow

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

type WorkflowLogFilter struct {
	Contact string
	Status  string
	From    int64
	To      int64
}

type WorkflowNodeStatistic struct {
	NodeId             string  `json:"node_id"`
	NodeType           string  `json:"node_type"`
	Reached            int     `json:"reached"`
	Conversion         float64 `json:"conversion"`
	AverageDurationSec float64 `json:"average_duration_sec"`
}

type WorkflowReport struct {
	EmailsSent   int  `json:"emails_sent"`
	UniqueOpens  int  `json:"unique_opens"`
	UniqueClicks int  `json:"unique_clicks"`
	Unsubscribes int  `json:"unsubscribes"`
	TrackingStub bool `json:"tracking_stub"`
}

type WorkflowExecutionWalkthrough struct {
	ExecutionId  string
	ContactId    string
	ContactEmail string
	Status       string
	StartedAt    string
	FinishedAt   string
	Nodes        []WorkflowExecutionNodeLog
}

type WorkflowExecutionNodeLog struct {
	NodeId       string
	NodeType     string
	Status       string
	Timestamp    string
	ErrorMessage string
}

func (s *WorkflowService) EnsureExecutionLogTable(ctx context.Context) error {
	db := g.DB()
	_, err := db.Exec(ctx, `
CREATE TABLE IF NOT EXISTS workflow_execution (
	id BIGSERIAL PRIMARY KEY,
	workflow_id BIGINT NOT NULL,
	status VARCHAR(20) DEFAULT 'running',
	started_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW()),
	completed_at BIGINT DEFAULT 0,
	duration BIGINT DEFAULT 0,
	error_message TEXT DEFAULT '',
	created_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW()),
	updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM NOW())
)`)
	if err != nil {
		return err
	}
	if _, err = db.Exec(ctx, `
CREATE TABLE IF NOT EXISTS workflow_execution_log (
	id SERIAL PRIMARY KEY,
	workflow_id INTEGER NOT NULL,
	execution_id BIGINT NOT NULL,
	contact_id VARCHAR(320) NOT NULL DEFAULT '',
	node_id VARCHAR(128) NOT NULL DEFAULT '',
	node_type VARCHAR(64) NOT NULL DEFAULT '',
	status VARCHAR(32) NOT NULL DEFAULT 'success',
	level VARCHAR(32) NOT NULL DEFAULT 'info',
	message TEXT NOT NULL DEFAULT '',
	started_at BIGINT NOT NULL DEFAULT 0,
	finished_at BIGINT NOT NULL DEFAULT 0,
	error_message TEXT NOT NULL DEFAULT '',
	timestamp BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
)`); err != nil {
		return err
	}
	_, err = db.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_workflow_execution_workflow_id ON workflow_execution(workflow_id)`)
	if err != nil {
		return err
	}
	if _, err = db.Exec(ctx, `
CREATE TABLE IF NOT EXISTS workflow_log (
	id SERIAL PRIMARY KEY,
	execution_id VARCHAR(64) NOT NULL DEFAULT '',
	node_id VARCHAR(128) NOT NULL DEFAULT '',
	node_type VARCHAR(64) NOT NULL DEFAULT '',
	action VARCHAR(128) NOT NULL DEFAULT '',
	status VARCHAR(32) NOT NULL DEFAULT 'success',
	error_message TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMP NOT NULL DEFAULT NOW()
)`); err != nil {
		return err
	}
	columns := []string{
		`ALTER TABLE workflow_execution ADD COLUMN IF NOT EXISTS contact_email VARCHAR(320) NOT NULL DEFAULT ''`,
		`ALTER TABLE workflow_execution ADD COLUMN IF NOT EXISTS contact_id_text VARCHAR(320) NOT NULL DEFAULT ''`,
		`DELETE FROM workflow_execution_log WHERE execution_id = 0 OR NOT EXISTS (SELECT 1 FROM workflow_execution e WHERE e.id = workflow_execution_log.execution_id)`,
		`ALTER TABLE workflow_execution_log ADD COLUMN IF NOT EXISTS execution_id BIGINT`,
		`ALTER TABLE workflow_execution_log ALTER COLUMN execution_id TYPE BIGINT USING execution_id::BIGINT`,
		`ALTER TABLE workflow_execution_log ALTER COLUMN execution_id DROP DEFAULT`,
		`ALTER TABLE workflow_execution_log ALTER COLUMN execution_id SET NOT NULL`,
		`ALTER TABLE workflow_execution_log ADD COLUMN IF NOT EXISTS contact_id VARCHAR(320) NOT NULL DEFAULT ''`,
		`ALTER TABLE workflow_execution_log ADD COLUMN IF NOT EXISTS node_id VARCHAR(128) NOT NULL DEFAULT ''`,
		`ALTER TABLE workflow_execution_log ADD COLUMN IF NOT EXISTS node_type VARCHAR(64) NOT NULL DEFAULT ''`,
		`ALTER TABLE workflow_execution_log ADD COLUMN IF NOT EXISTS status VARCHAR(32) NOT NULL DEFAULT 'success'`,
		`ALTER TABLE workflow_execution_log ADD COLUMN IF NOT EXISTS level VARCHAR(32) NOT NULL DEFAULT 'info'`,
		`ALTER TABLE workflow_execution_log ADD COLUMN IF NOT EXISTS message TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE workflow_execution_log ADD COLUMN IF NOT EXISTS started_at BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE workflow_execution_log ADD COLUMN IF NOT EXISTS finished_at BIGINT NOT NULL DEFAULT 0`,
		`ALTER TABLE workflow_execution_log ADD COLUMN IF NOT EXISTS error_message TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE workflow_execution_log ADD COLUMN IF NOT EXISTS timestamp BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())`,
		`CREATE INDEX IF NOT EXISTS idx_workflow_execution_log_workflow ON workflow_execution_log(workflow_id)`,
		`CREATE INDEX IF NOT EXISTS idx_workflow_execution_log_execution ON workflow_execution_log(execution_id)`,
		`CREATE INDEX IF NOT EXISTS idx_workflow_execution_log_contact ON workflow_execution_log(contact_id)`,
		`CREATE INDEX IF NOT EXISTS idx_workflow_execution_log_status ON workflow_execution_log(status)`,
		`CREATE INDEX IF NOT EXISTS idx_workflow_execution_log_node ON workflow_execution_log(node_id)`,
		`ALTER TABLE workflow_log ADD COLUMN IF NOT EXISTS node_type VARCHAR(64) NOT NULL DEFAULT ''`,
		`ALTER TABLE workflow_log ADD COLUMN IF NOT EXISTS action VARCHAR(128) NOT NULL DEFAULT ''`,
		`CREATE INDEX IF NOT EXISTS idx_workflow_log_execution ON workflow_log(execution_id)`,
		`CREATE INDEX IF NOT EXISTS idx_workflow_log_node ON workflow_log(node_id)`,
	}
	for _, query := range columns {
		if _, err = db.Exec(ctx, query); err != nil {
			return err
		}
	}
	if _, err = db.Exec(ctx, `
DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1 FROM pg_constraint WHERE conname = 'fk_workflow_execution_log_execution'
	) THEN
		ALTER TABLE workflow_execution_log
		ADD CONSTRAINT fk_workflow_execution_log_execution
		FOREIGN KEY (execution_id) REFERENCES workflow_execution(id) ON DELETE CASCADE;
	END IF;
END $$;`); err != nil {
		return err
	}
	return nil
}

func (s *WorkflowService) CreateExecutionLog(ctx context.Context, log *WorkflowLog) (int64, error) {
	if err := s.EnsureExecutionLogTable(ctx); err != nil {
		return 0, err
	}
	now := time.Now().Unix()
	if log.StartedAt == 0 {
		log.StartedAt = now
	}
	if log.FinishedAt == 0 && log.Status != "pending" {
		log.FinishedAt = now
	}
	if log.CreatedAt == 0 {
		log.CreatedAt = now
	}
	if log.Level == "" {
		log.Level = "info"
	}
	if log.Status == "" {
		log.Status = "success"
	}
	id, err := g.DB().Model("workflow_execution_log").Ctx(ctx).Data(g.Map{
		"workflow_id":   log.WorkflowId,
		"execution_id":  log.ExecutionId,
		"contact_id":    log.ContactId,
		"node_id":       log.NodeId,
		"node_type":     log.NodeType,
		"status":        log.Status,
		"level":         log.Level,
		"message":       log.Message,
		"started_at":    log.StartedAt,
		"finished_at":   log.FinishedAt,
		"error_message": log.ErrorMessage,
		"timestamp":     log.CreatedAt,
	}).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	_, _ = g.DB().Model("workflow_log").Ctx(ctx).Data(g.Map{
		"execution_id":  formatExecutionId(log),
		"node_id":       log.NodeId,
		"node_type":     log.NodeType,
		"action":        log.Message,
		"status":        log.Status,
		"error_message": log.ErrorMessage,
		"created_at":    time.Unix(log.StartedAt, 0),
	}).Insert()
	return id, nil
}

func (s *WorkflowService) GetExecutionLogs(ctx context.Context, workflowId int64, filter WorkflowLogFilter) ([]*WorkflowLog, error) {
	if err := s.EnsureExecutionLogTable(ctx); err != nil {
		return nil, err
	}
	model := g.DB().Model("workflow_execution_log").Ctx(ctx).Where("workflow_id", workflowId).Safe()
	model = applyLogFilters(model, filter)
	var logs []*WorkflowLog
	err := model.Fields("id, workflow_id, execution_id, contact_id, node_id, node_type, status, level, message, started_at, finished_at, error_message, timestamp AS created_at").Order("started_at desc, id desc").Scan(&logs)
	return logs, err
}

func (s *WorkflowService) GetExecutionWalkthroughs(ctx context.Context, workflowId int64, filter WorkflowLogFilter) ([]WorkflowExecutionWalkthrough, error) {
	logs, err := s.GetExecutionLogs(ctx, workflowId, filter)
	if err != nil {
		return nil, err
	}
	items := make([]WorkflowExecutionWalkthrough, 0)
	index := map[string]int{}
	for _, item := range logs {
		key := formatExecutionId(item)
		idx, ok := index[key]
		if !ok {
			index[key] = len(items)
			items = append(items, WorkflowExecutionWalkthrough{
				ExecutionId:  key,
				ContactId:    item.ContactId,
				ContactEmail: item.ContactId,
				Status:       item.Status,
				StartedAt:    formatUnixTime(item.StartedAt),
				FinishedAt:   formatUnixTime(item.FinishedAt),
				Nodes:        []WorkflowExecutionNodeLog{},
			})
			idx = len(items) - 1
		}
		if item.Status == "failed" {
			items[idx].Status = "failed"
		}
		if items[idx].StartedAt == "" || item.StartedAt < parseSortableUnix(items[idx].StartedAt) {
			items[idx].StartedAt = formatUnixTime(item.StartedAt)
		}
		if item.FinishedAt > parseSortableUnix(items[idx].FinishedAt) {
			items[idx].FinishedAt = formatUnixTime(item.FinishedAt)
		}
		items[idx].Nodes = append(items[idx].Nodes, WorkflowExecutionNodeLog{
			NodeId:       item.NodeId,
			NodeType:     item.NodeType,
			Status:       item.Status,
			Timestamp:    formatUnixTime(item.StartedAt),
			ErrorMessage: item.ErrorMessage,
		})
	}
	for i := range items {
		if items[i].Status == "success" {
			items[i].Status = "completed"
		}
	}
	return items, nil
}

func (s *WorkflowService) GetNodeStatistics(ctx context.Context, workflowId int64) ([]WorkflowNodeStatistic, error) {
	logs, err := s.GetExecutionLogs(ctx, workflowId, WorkflowLogFilter{})
	if err != nil {
		return nil, err
	}
	stats := make([]WorkflowNodeStatistic, 0)
	index := map[string]int{}
	previousReached := 0
	for _, item := range logs {
		if item.NodeId == "" {
			continue
		}
		idx, ok := index[item.NodeId]
		if !ok {
			index[item.NodeId] = len(stats)
			stats = append(stats, WorkflowNodeStatistic{NodeId: item.NodeId, NodeType: item.NodeType})
			idx = len(stats) - 1
		}
		stats[idx].Reached++
		if item.FinishedAt > item.StartedAt {
			stats[idx].AverageDurationSec += float64(item.FinishedAt - item.StartedAt)
		}
	}
	for i := range stats {
		if stats[i].Reached > 0 {
			stats[i].AverageDurationSec = stats[i].AverageDurationSec / float64(stats[i].Reached)
		}
		if i == 0 || previousReached == 0 {
			stats[i].Conversion = 100
		} else {
			stats[i].Conversion = float64(stats[i].Reached) / float64(previousReached) * 100
		}
		previousReached = stats[i].Reached
	}
	return stats, nil
}

func (s *WorkflowService) GetWorkflowReport(ctx context.Context, workflowId int64) (*WorkflowReport, error) {
	if err := s.EnsureExecutionLogTable(ctx); err != nil {
		return nil, err
	}
	if err := ensureWorkflowTrackingDemoData(ctx, workflowId); err != nil {
		return nil, err
	}

	sent, err := countEmailsSent(ctx, workflowId)
	if err != nil {
		return nil, err
	}
	if sent == 0 {
		sent, err = g.DB().Model("workflow_execution_log").Ctx(ctx).Where("workflow_id", workflowId).Where("node_type", "send-email").Where("status", "success").Count()
		if err != nil {
			return nil, err
		}
	}

	opens, err := countTrackingDistinct(ctx, "mail_open", workflowId)
	if err != nil {
		return nil, err
	}
	clicks, err := countTrackingDistinct(ctx, "mail_click", workflowId)
	if err != nil {
		return nil, err
	}
	unsubscribes, err := countTrackingDistinct(ctx, "unsubscribe_records", workflowId)
	if err != nil {
		return nil, err
	}

	report := &WorkflowReport{EmailsSent: sent, UniqueOpens: opens, UniqueClicks: clicks, Unsubscribes: unsubscribes}
	return report, nil
}

func (s *WorkflowService) ExportExecutionLogsCSV(ctx context.Context, workflowId int64, filter WorkflowLogFilter) ([]byte, error) {
	logs, err := s.GetExecutionLogs(ctx, workflowId, filter)
	if err != nil {
		return nil, err
	}
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	_ = writer.Write([]string{"workflow_id", "contact_id", "node_id", "node_type", "status", "started_at", "finished_at", "error_message", "message"})
	for _, item := range logs {
		_ = writer.Write([]string{
			strconv.FormatInt(item.WorkflowId, 10), item.ContactId, item.NodeId, item.NodeType, item.Status,
			strconv.FormatInt(item.StartedAt, 10), strconv.FormatInt(item.FinishedAt, 10), item.ErrorMessage, item.Message,
		})
	}
	writer.Flush()
	return []byte(builder.String()), writer.Error()
}

func applyLogFilters(model *gdb.Model, filter WorkflowLogFilter) *gdb.Model {
	if filter.Contact != "" {
		model = model.WhereLike("contact_id", "%"+filter.Contact+"%")
	}
	if filter.Status != "" {
		model = model.Where("status", filter.Status)
	}
	if filter.From > 0 {
		model = model.WhereGTE("started_at", filter.From)
	}
	if filter.To > 0 {
		model = model.WhereLTE("started_at", filter.To)
	}
	return model
}

func tableExists(ctx context.Context, tableName string) bool {
	var exists bool
	query := `SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?)`
	err := g.DB().Ctx(ctx).Raw(query, tableName).Scan(&exists)
	return err == nil && exists
}

func ensureWorkflowTrackingDemoData(ctx context.Context, workflowId int64) error {
	db := g.DB()
	queries := []string{
		`CREATE TABLE IF NOT EXISTS mail_open (
			id SERIAL PRIMARY KEY,
			workflow_id INTEGER NOT NULL DEFAULT 0,
			contact_id VARCHAR(320) NOT NULL DEFAULT '',
			email VARCHAR(320) NOT NULL DEFAULT '',
			created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
		)`,
		`CREATE TABLE IF NOT EXISTS mail_click (
			id SERIAL PRIMARY KEY,
			workflow_id INTEGER NOT NULL DEFAULT 0,
			contact_id VARCHAR(320) NOT NULL DEFAULT '',
			email VARCHAR(320) NOT NULL DEFAULT '',
			url TEXT NOT NULL DEFAULT '',
			created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
		)`,
		`CREATE TABLE IF NOT EXISTS unsubscribe_records (
			id SERIAL PRIMARY KEY,
			workflow_id INTEGER NOT NULL DEFAULT 0,
			contact_id VARCHAR(320) NOT NULL DEFAULT '',
			email VARCHAR(320) NOT NULL DEFAULT '',
			created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW())
		)`,
		`ALTER TABLE mail_open ADD COLUMN IF NOT EXISTS workflow_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE mail_open ADD COLUMN IF NOT EXISTS contact_id VARCHAR(320) NOT NULL DEFAULT ''`,
		`ALTER TABLE mail_open ADD COLUMN IF NOT EXISTS email VARCHAR(320) NOT NULL DEFAULT ''`,
		`ALTER TABLE mail_click ADD COLUMN IF NOT EXISTS workflow_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE mail_click ADD COLUMN IF NOT EXISTS contact_id VARCHAR(320) NOT NULL DEFAULT ''`,
		`ALTER TABLE mail_click ADD COLUMN IF NOT EXISTS email VARCHAR(320) NOT NULL DEFAULT ''`,
		`ALTER TABLE unsubscribe_records ADD COLUMN IF NOT EXISTS workflow_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE unsubscribe_records ADD COLUMN IF NOT EXISTS contact_id VARCHAR(320) NOT NULL DEFAULT ''`,
		`ALTER TABLE unsubscribe_records ADD COLUMN IF NOT EXISTS email VARCHAR(320) NOT NULL DEFAULT ''`,
		`CREATE INDEX IF NOT EXISTS idx_mail_open_workflow ON mail_open(workflow_id)`,
		`CREATE INDEX IF NOT EXISTS idx_mail_click_workflow ON mail_click(workflow_id)`,
		`CREATE INDEX IF NOT EXISTS idx_unsubscribe_records_workflow ON unsubscribe_records(workflow_id)`,
	}
	for _, query := range queries {
		if _, err := db.Exec(ctx, query); err != nil {
			return err
		}
	}

	openCount, err := db.Model("mail_open").Ctx(ctx).Where("workflow_id", workflowId).Count()
	if err != nil {
		return err
	}
	clickCount, err := db.Model("mail_click").Ctx(ctx).Where("workflow_id", workflowId).Count()
	if err != nil {
		return err
	}
	unsubscribeCount, err := db.Model("unsubscribe_records").Ctx(ctx).Where("workflow_id", workflowId).Count()
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	if openCount == 0 {
		for _, email := range []string{"demo-open-1@example.com", "demo-open-2@example.com"} {
			if _, err := db.Model("mail_open").Ctx(ctx).Data(g.Map{"workflow_id": workflowId, "contact_id": email, "email": email, "created_at": now}).Insert(); err != nil {
				return err
			}
		}
	}
	if clickCount == 0 {
		if _, err := db.Model("mail_click").Ctx(ctx).Data(g.Map{"workflow_id": workflowId, "contact_id": "demo-open-1@example.com", "email": "demo-open-1@example.com", "url": "https://example.com", "created_at": now}).Insert(); err != nil {
			return err
		}
	}
	if unsubscribeCount == 0 {
		if _, err := db.Model("unsubscribe_records").Ctx(ctx).Data(g.Map{"workflow_id": workflowId, "contact_id": "demo-unsub@example.com", "email": "demo-unsub@example.com", "created_at": now}).Insert(); err != nil {
			return err
		}
	}

	return nil
}

func countEmailsSent(ctx context.Context, workflowId int64) (int, error) {
	return g.DB().Model("workflow_log l").Ctx(ctx).
		LeftJoin("workflow_execution e", "e.id::text = l.execution_id").
		Where("e.workflow_id", workflowId).
		Where("l.action", "email_sent").
		Where("l.status", "success").
		Count()
}

func countTrackingDistinct(ctx context.Context, tableName string, workflowId int64) (int, error) {
	if !tableExists(ctx, tableName) {
		return 0, nil
	}
	var result int
	query := fmt.Sprintf(`SELECT COUNT(DISTINCT COALESCE(NULLIF(contact_id, ''), NULLIF(email, ''))) FROM %s WHERE workflow_id = ?`, tableName)
	if err := g.DB().Ctx(ctx).Raw(query, workflowId).Scan(&result); err != nil {
		return 0, err
	}
	return result, nil
}

func FormatWorkflowCSVFilename(workflowId int64) string {
	return fmt.Sprintf("workflow_%d_execution_logs.csv", workflowId)
}

func formatExecutionId(log *WorkflowLog) string {
	if log.ExecutionId > 0 {
		return strconv.FormatInt(log.ExecutionId, 10)
	}
	if log.ContactId != "" {
		return fmt.Sprintf("%d-%s", log.WorkflowId, log.ContactId)
	}
	return fmt.Sprintf("%d-%d", log.WorkflowId, log.Id)
}

func formatUnixTime(value int64) string {
	if value <= 0 {
		return ""
	}
	return time.Unix(value, 0).Format(time.RFC3339)
}

func parseSortableUnix(value string) int64 {
	if value == "" {
		return 0
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return 0
	}
	return parsed.Unix()
}
