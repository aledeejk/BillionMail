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
	g.Meta `path:"/workflow/{id}/execute" method:"post" tags:"Workflow" summary:"Execute workflow"`
	Id      string `json:"id" v:"required"`
	Trigger string `json:"trigger"`
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
	ActiveContacts  int     `json:"active_contacts"`
	EmailsSent      int     `json:"emails_sent"`
	ConversionRate  float64 `json:"conversion_rate"`
}

// Editor API types
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
	Id          string                  `json:"id" in:"path" v:"required" dc:"Workflow ID"`
	Nodes       []*WorkflowNodeItem    `json:"nodes"`
	Connections []*WorkflowConnection  `json:"connections"`
}

type UpdateWorkflowEditorRes struct{}

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
	Id     string `json:"id" dc:"Connection ID"`
	Source string `json:"source" dc:"Source node ID"`
	Target string `json:"target" dc:"Target node ID"`
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
