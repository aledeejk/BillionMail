package entity

import (
	"database/sql/driver"
	"encoding/json"
)

// WorkflowConnection represents a connection between workflow nodes.
type WorkflowConnection struct {
	Id         string                 `json:"id" db:"id"`
	WorkflowId string                 `json:"workflow_id" db:"workflow_id"`
	FromNodeId string                 `json:"from_node_id" db:"from_node_id"`
	ToNodeId   string                 `json:"to_node_id" db:"to_node_id"`
	Condition  map[string]interface{} `json:"condition" db:"condition"`
	CreatedAt  string                 `json:"created_at" db:"created_at"`
}

// Value implements the driver.Valuer interface for JSONB
func (w WorkflowConnection) Value() (driver.Value, error) {
	return json.Marshal(w.Condition)
}