package workflow

import (
	"context"
	"encoding/json"
	"time"

	v1 "billionmail-core/api/workflow/v1"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

const (
	editorWorkflowTable   = "workflow"
	editorNodeTable       = "workflow_node"
	editorConnectionTable = "workflow_connection"
)

type ServiceEditor struct{}

var (
	editorServiceInstance *ServiceEditor
)

// WorkflowEditor returns the singleton instance of ServiceEditor.
func WorkflowEditor() *ServiceEditor {
	if editorServiceInstance == nil {
		editorServiceInstance = &ServiceEditor{}
	}
	return editorServiceInstance
}

// GetEditorData retrieves workflow editor data including nodes and connections.
func (s *ServiceEditor) GetEditorData(ctx context.Context, workflowId string) (*v1.GetWorkflowEditorRes, error) {
	id := gconv.Int64(workflowId)
	if id <= 0 {
		return nil, gerror.New("workflow id is required")
	}

	if err := s.ensureEditorTables(ctx, g.DB()); err != nil {
		return nil, err
	}

	workflowRecord, err := g.DB().Model(editorWorkflowTable).Ctx(ctx).Where("id", id).One()
	if err != nil {
		return nil, err
	}
	if workflowRecord.IsEmpty() {
		return nil, gerror.Newf("workflow %d not found", id)
	}

	nodeRecords, err := g.DB().Model(editorNodeTable).Ctx(ctx).
		Where("workflow_id", id).
		Order("created_at asc").
		All()
	if err != nil {
		return nil, err
	}

	nodes := make([]*v1.WorkflowNodeItem, 0, len(nodeRecords))
	for _, record := range nodeRecords {
		config := map[string]interface{}{}
		if raw := record["config"].String(); raw != "" {
			_ = json.Unmarshal([]byte(raw), &config)
		}

		nodes = append(nodes, &v1.WorkflowNodeItem{
			Id:        record["id"].String(),
			Type:      record["type"].String(),
			Config:    config,
			PositionX: record["position_x"].Float64(),
			PositionY: record["position_y"].Float64(),
		})
	}

	connectionRecords, err := g.DB().Model(editorConnectionTable).Ctx(ctx).
		Where("workflow_id", id).
		Order("created_at asc").
		All()
	if err != nil {
		return nil, err
	}

	connections := make([]*v1.WorkflowConnection, 0, len(connectionRecords))
	for _, record := range connectionRecords {
		connections = append(connections, &v1.WorkflowConnection{
			Id:     record["id"].String(),
			Source: record["source"].String(),
			Target: record["target"].String(),
		})
	}

	return &v1.GetWorkflowEditorRes{
		Workflow: &v1.WorkflowItem{
			Id:          workflowRecord["id"].String(),
			Name:        workflowRecord["name"].String(),
			Description: workflowRecord["description"].String(),
			IsActive:    workflowRecord["status"].Int() == 1,
			Version:     workflowRecord["version"].Int(),
			CreatedAt:   workflowRecord["created_at"].String(),
			UpdatedAt:   workflowRecord["updated_at"].String(),
		},
		Nodes:       nodes,
		Connections: connections,
	}, nil
}

// UpdateEditorData replaces workflow editor nodes and connections transactionally.
func (s *ServiceEditor) UpdateEditorData(ctx context.Context, workflowId string, nodes []*v1.WorkflowNodeItem, connections []*v1.WorkflowConnection) error {
	id := gconv.Int64(workflowId)
	if id <= 0 {
		return gerror.New("workflow id is required")
	}

	db := g.DB()
	if err := s.ensureEditorTables(ctx, db); err != nil {
		return err
	}

	exists, err := db.Model(editorWorkflowTable).Ctx(ctx).Where("id", id).Exist()
	if err != nil {
		return err
	}
	if !exists {
		return gerror.Newf("workflow %d not found", id)
	}

	now := time.Now().Unix()

	return db.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if _, err := tx.Model(editorConnectionTable).Where("workflow_id", id).Delete(); err != nil {
			return err
		}
		if _, err := tx.Model(editorNodeTable).Where("workflow_id", id).Delete(); err != nil {
			return err
		}

		for _, node := range nodes {
			if node == nil || node.Id == "" {
				continue
			}

			configJSON, err := json.Marshal(node.Config)
			if err != nil {
				return err
			}

			if _, err := tx.Model(editorNodeTable).Data(g.Map{
				"id":          node.Id,
				"workflow_id": id,
				"type":        node.Type,
				"config":      string(configJSON),
				"position_x":  node.PositionX,
				"position_y":  node.PositionY,
				"created_at":  now,
				"updated_at":  now,
			}).Insert(); err != nil {
				return err
			}
		}

		for _, connection := range connections {
			if connection == nil || connection.Id == "" {
				continue
			}

			if _, err := tx.Model(editorConnectionTable).Data(g.Map{
				"id":          connection.Id,
				"workflow_id": id,
				"source":      connection.Source,
				"target":      connection.Target,
				"created_at":  now,
				"updated_at":  now,
			}).Insert(); err != nil {
				return err
			}
		}

		_, err := tx.Model(editorWorkflowTable).Where("id", id).Data(g.Map{
			"updated_at": now,
		}).Update()
		return err
	})
}

func (s *ServiceEditor) ensureEditorTables(ctx context.Context, db gdb.DB) error {
	_, err := db.Exec(ctx, `
CREATE TABLE IF NOT EXISTS workflow_node (
	id VARCHAR(128) PRIMARY KEY,
	workflow_id INTEGER NOT NULL REFERENCES workflow(id) ON DELETE CASCADE,
	type VARCHAR(50) NOT NULL,
	config JSONB NOT NULL DEFAULT '{}'::jsonb,
	position_x DOUBLE PRECISION NOT NULL DEFAULT 0,
	position_y DOUBLE PRECISION NOT NULL DEFAULT 0,
	created_at BIGINT NOT NULL DEFAULT 0,
	updated_at BIGINT NOT NULL DEFAULT 0
)`)
	if err != nil {
		return err
	}

	_, err = db.Exec(ctx, `
CREATE TABLE IF NOT EXISTS workflow_connection (
	id VARCHAR(128) PRIMARY KEY,
	workflow_id INTEGER NOT NULL REFERENCES workflow(id) ON DELETE CASCADE,
	source VARCHAR(128) NOT NULL,
	target VARCHAR(128) NOT NULL,
	created_at BIGINT NOT NULL DEFAULT 0,
	updated_at BIGINT NOT NULL DEFAULT 0
)`)
	if err != nil {
		return err
	}

	if _, err = db.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_workflow_node_workflow_id ON workflow_node(workflow_id)`); err != nil {
		return err
	}
	if _, err = db.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_workflow_connection_workflow_id ON workflow_connection(workflow_id)`); err != nil {
		return err
	}

	return nil
}
