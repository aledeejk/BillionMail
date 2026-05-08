package entity

import (
	"database/sql/driver"
	"encoding/json"
)

// WorkflowNode represents a node in a workflow.
type WorkflowNode struct {
	Id        string                 `json:"id" db:"id"`
	WorkflowId string                `json:"workflow_id" db:"workflow_id"`
	NodeType  string                 `json:"node_type" db:"node_type"`
	Config    map[string]interface{} `json:"config" db:"config"`
	PositionX float64                `json:"position_x" db:"position_x"`
	PositionY float64                `json:"position_y" db:"position_y"`
	CreatedAt string                 `json:"created_at" db:"created_at"`
	UpdatedAt string                 `json:"updated_at" db:"updated_at"`
}

// Value implements the driver.Valuer interface for JSONB
func (w WorkflowNode) Value() (driver.Value, error) {
	return json.Marshal(w.Config)
}