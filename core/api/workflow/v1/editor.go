package v1

import "github.com/gogf/gf/v2/frame/g"

// GetWorkflowEditorReq 获取工作流编辑器数据请求
type GetWorkflowEditorReq struct {
	g.Meta `path:"/workflow/{id}/editor" method:"get" tags:"Workflow" summary:"Get workflow editor data"`
	Id     string `json:"id" v:"required" dc:"Workflow ID"`
}

// GetWorkflowEditorRes 获取工作流编辑器数据响应
type GetWorkflowEditorRes struct {
	Workflow    *WorkflowItem         `json:"workflow"`
	Nodes       []*WorkflowNodeItem   `json:"nodes"`
	Connections []*WorkflowConnection `json:"connections"`
}

// UpdateWorkflowEditorReq 更新工作流编辑器数据请求
type UpdateWorkflowEditorReq struct {
	g.Meta      `path:"/workflow/{id}/editor" method:"put" tags:"Workflow" summary:"Update workflow editor data"`
	Id          string                  `json:"id" v:"required" dc:"Workflow ID"`
	Nodes       []*WorkflowNodeItem    `json:"nodes"`
	Connections []*WorkflowConnection  `json:"connections"`
}

// UpdateWorkflowEditorRes 更新工作流编辑器数据响应
type UpdateWorkflowEditorRes struct{}

// WorkflowNodeItem 工作流节点项
type WorkflowNodeItem struct {
	Id        string                 `json:"id" dc:"Node ID"`
	Type      string                 `json:"type" dc:"Node type"`
	Config    map[string]interface{} `json:"config" dc:"Node configuration"`
	PositionX float64                `json:"position_x" dc:"X position"`
	PositionY float64                `json:"position_y" dc:"Y position"`
}

// WorkflowConnection 工作流连接
type WorkflowConnection struct {
	Id     string `json:"id" dc:"Connection ID"`
	Source string `json:"source" dc:"Source node ID"`
	Target string `json:"target" dc:"Target node ID"`
}