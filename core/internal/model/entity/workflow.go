package entity

// Workflow represents a marketing automation workflow.
type Workflow struct {
	Id        string `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	IsActive  bool   `json:"is_active" db:"is_active"`
	Version   int    `json:"version" db:"version"`
	CreatedAt string `json:"created_at" db:"created_at"`
	UpdatedAt string `json:"updated_at" db:"updated_at"`
}