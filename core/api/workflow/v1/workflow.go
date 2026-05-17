package v1

import (
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

type CreateWorkflowReq struct {
	g.Meta      `path:"/workflow" method:"post" tags:"Workflow" summary:"Create workflow"`
	Name        string `json:"name" v:"required"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type UpdateWorkflowReq struct {
	g.Meta      `path:"/workflow/{id}" method:"put" tags:"Workflow" summary:"Update workflow"`
	Id          string `json:"id" in:"path"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type GetWorkflowListReq struct {
	g.Meta `path:"/workflow" method:"get" tags:"Workflow" summary:"List workflows"`
	Page   int    `json:"page" d:"1"`
	Limit  int    `json:"limit" d:"20"`
	Search string `json:"search"`
}

type GetWorkflowReq struct {
	g.Meta `path:"/workflow/{id}" method:"get" tags:"Workflow" summary:"Get workflow"`
	Id     string `json:"id" v:"required"`
}

type DuplicateWorkflowReq struct {
	g.Meta `path:"/workflow/{id}/duplicate" method:"post" tags:"Workflow" summary:"Duplicate workflow"`
	Id     string `json:"id" v:"required"`
	Name   string `json:"name" v:"required"`
}

type ToggleWorkflowReq struct {
	g.Meta `path:"/workflow/{id}/toggle" method:"post" tags:"Workflow" summary:"Toggle workflow"`
	Id     string `json:"id" v:"required"`
}

type DeleteWorkflowReq struct {
	g.Meta `path:"/workflow/{id}" method:"delete" tags:"Workflow" summary:"Delete workflow"`
	Id     string `json:"id" v:"required"`
}

type GetWorkflowStatsReq struct {
	g.Meta `path:"/workflow/{id}/stats" method:"get" tags:"Workflow" summary:"Get workflow stats"`
	Id     string `json:"id" v:"required"`
}

type GetWorkflowVersionsReq struct {
	g.Meta `path:"/workflow/{id}/versions" method:"get" tags:"Workflow" summary:"Get workflow versions"`
	Id     string `json:"id" v:"required"`
	Page   int    `json:"page" d:"1"`
	Limit  int    `json:"limit" d:"20"`
}

type GetWorkflowExecutionsReq struct {
	g.Meta `path:"/workflow/{id}/executions" method:"get" tags:"Workflow" summary:"Get workflow executions"`
	Id     string `json:"id" v:"required"`
	Page   int    `json:"page" d:"1"`
	Limit  int    `json:"limit" d:"20"`
}

type ExecuteWorkflowReq struct {
	g.Meta       `path:"/workflow/{id}/execute" method:"post" tags:"Workflow" summary:"Execute workflow"`
	Id           string                 `json:"id" v:"required"`
	Trigger      string                 `json:"trigger"`
	ContactId    int64                  `json:"contact_id"`
	ContactEmail string                 `json:"contact_email"`
	InputData    map[string]interface{} `json:"input_data"`
}

type CreateWorkflowExecutionLogReq struct {
	g.Meta       `path:"/workflow/{id}/execution-log" method:"post" tags:"Workflow" summary:"Create workflow execution log"`
	Id           string `json:"id" in:"path" v:"required"`
	ExecutionId  int64  `json:"execution_id"`
	ContactId    string `json:"contact_id"`
	NodeId       string `json:"node_id"`
	NodeType     string `json:"node_type"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	StartedAt    int64  `json:"started_at"`
	FinishedAt   int64  `json:"finished_at"`
	ErrorMessage string `json:"error_message"`
}

type GetWorkflowExecutionLogsReq struct {
	g.Meta  `path:"/workflow/{id}/execution-log" method:"get" tags:"Workflow" summary:"Get workflow execution logs"`
	Id      string `json:"id" in:"path" v:"required"`
	Contact string `json:"contact"`
	Status  string `json:"status"`
	From    int64  `json:"from"`
	To      int64  `json:"to"`
}

type GetWorkflowNodeStatsReq struct {
	g.Meta `path:"/workflow/{id}/node-stats" method:"get" tags:"Workflow" summary:"Get workflow node stats"`
	Id     string `json:"id" in:"path" v:"required"`
}

type GetWorkflowReportReq struct {
	g.Meta `path:"/workflow/{id}/report" method:"get" tags:"Workflow" summary:"Get workflow report"`
	Id     string `json:"id" in:"path" v:"required"`
}

type ExportWorkflowExecutionLogsReq struct {
	g.Meta  `path:"/workflow/{id}/execution-log/export" method:"get" tags:"Workflow" summary:"Export workflow execution logs"`
	Id      string `json:"id" in:"path" v:"required"`
	Contact string `json:"contact"`
	Status  string `json:"status"`
	From    int64  `json:"from"`
	To      int64  `json:"to"`
}

type GetExecutionLogReq struct {
	g.Meta  `path:"/workflow/{id}/execution-log" method:"get" tags:"Workflow" summary:"Get workflow execution log"`
	Id      int64  `json:"id" in:"path"`
	Contact string `json:"contact"`
	Status  string `json:"status"`
}

type GetReportReq struct {
	g.Meta `path:"/workflow/{id}/report" method:"get" tags:"Workflow" summary:"Get workflow report"`
	Id     int64 `json:"id" in:"path"`
}

type GetNodeStatsReq struct {
	g.Meta `path:"/workflow/{id}/node-stats" method:"get" tags:"Workflow" summary:"Get workflow node analytics"`
	Id     int64 `json:"id" in:"path"`
}

type ExportExecutionLogReq struct {
	g.Meta  `path:"/workflow/{id}/export" method:"get" tags:"Workflow" summary:"Export workflow execution log"`
	Id      int64  `json:"id" in:"path"`
	Contact string `json:"contact"`
	Status  string `json:"status"`
}

type DuplicateWorkflowRes struct {
	Workflow WorkflowRes `json:"workflow"`
}

type WorkflowRes struct {
	Id          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WorkflowListRes struct {
	List  []WorkflowRes `json:"list"`
	Total int           `json:"total"`
}

type WorkflowStatsRes struct {
	ActiveContacts int     `json:"active_contacts"`
	EmailsSent     int     `json:"emails_sent"`
	ConversionRate float64 `json:"conversion_rate"`
}

type GetWorkflowEditorReq struct {
	g.Meta `path:"/workflow/{id}/editor" method:"get" tags:"Workflow" summary:"Get workflow editor data"`
	Id     string `json:"id" in:"path" v:"required" dc:"Workflow ID"`
}

type GetWorkflowEditorRes struct {
	Workflow    *WorkflowItem         `json:"workflow"`
	Nodes       []*WorkflowNodeItem   `json:"nodes"`
	Connections []*WorkflowConnection `json:"connections"`
}

type UpdateWorkflowEditorReq struct {
	g.Meta      `path:"/workflow/{id}/editor" method:"put" tags:"Workflow" summary:"Update workflow editor data"`
	Id          string                `json:"id" in:"path" v:"required" dc:"Workflow ID"`
	Nodes       []*WorkflowNodeItem   `json:"nodes"`
	Connections []*WorkflowConnection `json:"connections"`
}

type UpdateWorkflowEditorRes struct{}

type RollbackWorkflowReq struct {
	g.Meta  `path:"/workflow/{id}/rollback/{version}" method:"post" tags:"Workflow" summary:"Rollback workflow editor data"`
	Id      string `json:"id" in:"path" v:"required" dc:"Workflow ID"`
	Version int    `json:"version" in:"path" v:"required" dc:"Version number"`
}

type WorkflowItem struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
	Version     int    `json:"version"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type WorkflowNodeItem struct {
	Id        string                 `json:"id" dc:"Node ID"`
	Type      string                 `json:"type" dc:"Node type"`
	Config    map[string]interface{} `json:"config" dc:"Node configuration"`
	PositionX float64                `json:"position_x" dc:"X position"`
	PositionY float64                `json:"position_y" dc:"Y position"`
}

type WorkflowConnection struct {
	Id        string `json:"id" dc:"Connection ID"`
	Source    string `json:"source" dc:"Source node ID"`
	Target    string `json:"target" dc:"Target node ID"`
	Condition string `json:"condition" dc:"Transition condition"`
}

type WorkflowVersionRes struct {
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	Status    int       `json:"status"`
}

type WorkflowExecutionRes struct {
	Id           string    `json:"id"`
	WorkflowId   string    `json:"workflow_id"`
	Version      int       `json:"version"`
	Status       int       `json:"status"`
	Trigger      string    `json:"trigger"`
	StartedAt    time.Time `json:"started_at"`
	CompletedAt  time.Time `json:"completed_at"`
	ErrorMessage string    `json:"error_message"`
}

type WorkflowExecutionLogRes struct {
	Id           string `json:"id"`
	WorkflowId   string `json:"workflow_id"`
	ExecutionId  int64  `json:"execution_id"`
	ContactId    string `json:"contact_id"`
	NodeId       string `json:"node_id"`
	NodeType     string `json:"node_type"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	StartedAt    int64  `json:"started_at"`
	FinishedAt   int64  `json:"finished_at"`
	ErrorMessage string `json:"error_message"`
}

type WorkflowExecutionLogListRes struct {
	List  []*WorkflowExecutionWalkthroughRes `json:"list"`
	Total int                                `json:"total"`
}

type WorkflowExecutionWalkthroughRes struct {
	ExecutionId  string                         `json:"execution_id"`
	ContactId    string                         `json:"contact_id"`
	ContactEmail string                         `json:"contact_email"`
	Status       string                         `json:"status"`
	StartedAt    string                         `json:"started_at"`
	FinishedAt   string                         `json:"finished_at"`
	Nodes        []*WorkflowExecutionNodeLogRes `json:"nodes"`
}

type WorkflowExecutionNodeLogRes struct {
	NodeId       string `json:"node_id"`
	NodeType     string `json:"node_type"`
	Status       string `json:"status"`
	Timestamp    string `json:"timestamp"`
	ErrorMessage string `json:"error_message"`
}

type ExecutionLogItem struct {
	ExecutionId  string     `json:"execution_id"`
	ContactId    string     `json:"contact_id"`
	ContactEmail string     `json:"contact_email"`
	Status       string     `json:"status"`
	StartedAt    time.Time  `json:"started_at"`
	FinishedAt   *time.Time `json:"finished_at"`
	Nodes        []LogNode  `json:"nodes"`
}

type LogNode struct {
	NodeId    string    `json:"node_id"`
	NodeType  string    `json:"node_type"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type GetExecutionLogRes struct {
	List  []ExecutionLogItem `json:"list"`
	Total int                `json:"total"`
}

type GetReportRes struct {
	Sent         int `json:"sent"`
	UniqueOpens  int `json:"unique_opens"`
	UniqueClicks int `json:"unique_clicks"`
	Unsubscribes int `json:"unsubscribes"`
}

type NodeStat struct {
	NodeId         string  `json:"node_id"`
	NodeType       string  `json:"node_type"`
	EnteredCount   int     `json:"entered_count"`
	CompletedCount int     `json:"completed_count"`
	ConversionRate float64 `json:"conversion_rate"`
	AvgDurationMs  int64   `json:"avg_duration_ms"`
}

type WorkflowNodeStatsRes struct {
	NodeId             string  `json:"node_id"`
	NodeType           string  `json:"node_type"`
	Reached            int     `json:"reached"`
	Conversion         float64 `json:"conversion"`
	AverageDurationSec float64 `json:"average_duration_sec"`
}

type WorkflowReportRes struct {
	Sent         int  `json:"sent"`
	EmailsSent   int  `json:"emails_sent"`
	UniqueOpens  int  `json:"unique_opens"`
	UniqueClicks int  `json:"unique_clicks"`
	Unsubscribes int  `json:"unsubscribes"`
	TrackingStub bool `json:"tracking_stub"`
}
