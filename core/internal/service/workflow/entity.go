package workflow

// Workflow represents a marketing automation workflow.
type Workflow struct {
	Id          int64                  `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      int                    `json:"status"` // 0: disabled, 1: enabled
	Version     int                    `json:"version"`
	Trigger     string                 `json:"trigger"`
	Nodes       []*WorkflowNode        `json:"nodes"`
	Connections []*WorkflowConnection  `json:"connections"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   int64                  `json:"created_at"`
	UpdatedAt   int64                  `json:"updated_at"`
}

// WorkflowNode describes a single node in an automation workflow.
type WorkflowNode struct {
	Id        string                 `json:"id"`
	Type      string                 `json:"type"`
	Name      string                 `json:"name"`
	Config    map[string]interface{} `json:"config,omitempty"`
	PositionX int                    `json:"position_x"`
	PositionY int                    `json:"position_y"`
	CreatedAt int64                  `json:"created_at"`
	UpdatedAt int64                  `json:"updated_at"`
}

// WorkflowConnection defines a connection between workflow nodes.
type WorkflowConnection struct {
	Id        string                 `json:"id"`
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Condition string                 `json:"condition,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt int64                  `json:"created_at"`
	UpdatedAt int64                  `json:"updated_at"`
}

// WorkflowVersion stores immutable workflow definitions for rollback and history.
type WorkflowVersion struct {
	Id          int64  `json:"id"`
	WorkflowId  int64  `json:"workflow_id"`
	Version     int    `json:"version"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      int    `json:"status"`
	Trigger     string `json:"trigger"`
	Definition  string `json:"definition"`
	CreatedAt   int64  `json:"created_at"`
}

// WorkflowExecution records a workflow run attempt.
type WorkflowExecution struct {
	Id          int64  `json:"id"`
	WorkflowId  int64  `json:"workflow_id"`
	Version     int    `json:"version"`
	Status      int    `json:"status"` // 0: pending, 1: running, 2: success, 3: failed
	Trigger     string `json:"trigger"`
	StartedAt   int64  `json:"started_at"`
	CompletedAt int64  `json:"completed_at"`
	Duration    int64  `json:"duration"`
	Result      string `json:"result"`
	Error       string `json:"error,omitempty"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// WorkflowLog stores runtime events for workflow executions.
type WorkflowLog struct {
	Id           int64  `json:"id"`
	WorkflowId   int64  `json:"workflow_id"`
	ExecutionId  int64  `json:"execution_id"`
	ContactId    string `json:"contact_id"`
	NodeId       string `json:"node_id"`
	NodeType     string `json:"node_type"`
	Status       string `json:"status"`
	Level        string `json:"level"`
	Message      string `json:"message"`
	StartedAt    int64  `json:"started_at"`
	FinishedAt   int64  `json:"finished_at"`
	ErrorMessage string `json:"error_message"`
	CreatedAt    int64  `json:"created_at"`
}
