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
