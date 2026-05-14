package workflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	v1 "billionmail-core/api/workflow/v1"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/google/uuid"
)

const (
	editorWorkflowTable   = "workflow"
	editorNodeTable       = "workflow_node"
	editorConnectionTable = "workflow_connection"
)

var (
	ErrCurrentWorkflowVersion  = errors.New("The current version cannot be deleted")
	ErrWorkflowVersionNotFound = errors.New("workflow version not found")
)

type ServiceEditor struct{}

type editorSnapshot struct {
	Nodes       []*v1.WorkflowNodeItem   `json:"nodes"`
	Connections []*v1.WorkflowConnection `json:"connections"`
}

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
			Id:        record["id"].String(),
			Source:    record["source"].String(),
			Target:    record["target"].String(),
			Condition: record["condition"].String(),
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

func (s *ServiceEditor) DeleteWorkflowVersion(ctx context.Context, workflowId string, version int) error {
	id := gconv.Int64(workflowId)
	if id <= 0 {
		return gerror.New("workflow id is required")
	}
	if version <= 0 {
		return gerror.New("version is required")
	}

	workflowRecord, err := g.DB().Model(editorWorkflowTable).Ctx(ctx).
		Fields("id, version").
		Where("id", id).
		One()
	if err != nil {
		return err
	}
	if workflowRecord.IsEmpty() {
		return ErrWorkflowVersionNotFound
	}
	if workflowRecord["version"].Int() == version {
		return ErrCurrentWorkflowVersion
	}

	exists, err := g.DB().Model("workflow_version").Ctx(ctx).
		Where("workflow_id", id).
		Where("version", version).
		Count()
	if err != nil {
		return err
	}
	if exists == 0 {
		return ErrWorkflowVersionNotFound
	}

	_, err = g.DB().Model("workflow_version").Ctx(ctx).
		Where("workflow_id", id).
		Where("version", version).
		Delete()
	return err
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
				"condition":   connection.Condition,
				"created_at":  now,
				"updated_at":  now,
			}).Insert(); err != nil {
				return err
			}
		}

		_, err := tx.Model(editorWorkflowTable).Where("id", id).Data(g.Map{
			"version":    gdb.Raw("version + 1"),
			"updated_at": now,
		}).Update()
		if err != nil {
			return err
		}

		var workflowRecord struct {
			Version int `json:"version"`
		}
		if err := tx.Model(editorWorkflowTable).Fields("version").Where("id", id).Scan(&workflowRecord); err != nil {
			return err
		}
		return s.createSnapshotVersion(ctx, tx, id, workflowRecord.Version, nodes, connections, now)
	})
}

// RollbackEditorData restores workflow nodes and connections from a specific version snapshot.
func (s *ServiceEditor) RollbackEditorData(ctx context.Context, workflowId string, version int) (*v1.GetWorkflowEditorRes, error) {
	fmt.Println("[ROLLBACK] no version increment")

	id := gconv.Int64(workflowId)
	if id <= 0 {
		return nil, gerror.New("workflow id is required")
	}
	if version <= 0 {
		return nil, gerror.New("version is required")
	}

	db := g.DB()
	if err := s.ensureEditorTables(ctx, db); err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	var restoredNodes []*v1.WorkflowNodeItem
	var restoredConnections []*v1.WorkflowConnection

	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var record struct {
		Snapshot string `json:"snapshot"`
	}
	if err := tx.Model("workflow_version").
		Fields("COALESCE(NULLIF(snapshot::text, '{}'), NULLIF(content, ''), '{}') AS snapshot").
		Where("workflow_id", id).
		Where("version", version).
		Order("id desc").
		Limit(1).
		Scan(&record); err != nil {
		return nil, err
	}
	if record.Snapshot == "" {
		return nil, gerror.Newf("workflow version %d not found", version)
	}
	fmt.Println("[ROLLBACK] snapshot loaded without workflow update")

	var snapshot struct {
		Nodes []struct {
			Id        string                 `json:"id"`
			Type      string                 `json:"type"`
			Config    map[string]interface{} `json:"config"`
			PositionX float64                `json:"position_x"`
			PositionY float64                `json:"position_y"`
		} `json:"nodes"`
		Connections []struct {
			Id        string `json:"id"`
			Source    string `json:"source"`
			Target    string `json:"target"`
			Condition string `json:"condition"`
		} `json:"connections"`
	}
	if err := json.Unmarshal([]byte(record.Snapshot), &snapshot); err != nil {
		return nil, err
	}
	if len(snapshot.Nodes) == 0 {
		return nil, gerror.Newf("workflow version %d snapshot has no nodes", version)
	}

	connectionDeleteResult, err := tx.Exec("DELETE FROM workflow_connection WHERE workflow_id = ?", id)
	if err != nil {
		return nil, err
	}
	nodeDeleteResult, err := tx.Exec("DELETE FROM workflow_node WHERE workflow_id = ?", id)
	if err != nil {
		return nil, err
	}
	deletedConnectionCount, _ := connectionDeleteResult.RowsAffected()
	deletedNodeCount, _ := nodeDeleteResult.RowsAffected()
	fmt.Println("[ROLLBACK] old nodes and connections deleted without workflow update; deleted nodes:", deletedNodeCount, "connections:", deletedConnectionCount)

	nodeIdMap := make(map[string]string, len(snapshot.Nodes))
	insertedNodeCount := 0
	restoredNodes = make([]*v1.WorkflowNodeItem, 0, len(snapshot.Nodes))
	for _, node := range snapshot.Nodes {
		if node.Id == "" {
			continue
		}

		newNodeId := uuid.NewString()
		nodeIdMap[node.Id] = newNodeId

		configJSON, err := json.Marshal(node.Config)
		if err != nil {
			return nil, err
		}

		if _, err := tx.Model(editorNodeTable).Data(g.Map{
			"id":          newNodeId,
			"workflow_id": id,
			"type":        node.Type,
			"config":      string(configJSON),
			"position_x":  node.PositionX,
			"position_y":  node.PositionY,
			"created_at":  now,
			"updated_at":  now,
		}).Insert(); err != nil {
			return nil, err
		}

		restoredNodes = append(restoredNodes, &v1.WorkflowNodeItem{
			Id:        newNodeId,
			Type:      node.Type,
			Config:    node.Config,
			PositionX: node.PositionX,
			PositionY: node.PositionY,
		})
		insertedNodeCount++
	}
	fmt.Println("[ROLLBACK] nodes inserted without workflow update:", insertedNodeCount)

	restoredConnections = make([]*v1.WorkflowConnection, 0, len(snapshot.Connections))
	insertedConnectionCount := 0
	for _, connection := range snapshot.Connections {
		if connection.Source == "" || connection.Target == "" {
			continue
		}

		newSource, sourceOk := nodeIdMap[connection.Source]
		newTarget, targetOk := nodeIdMap[connection.Target]
		if !sourceOk || !targetOk {
			continue
		}

		newConnectionId := uuid.NewString()
		if _, err := tx.Model(editorConnectionTable).Data(g.Map{
			"id":          newConnectionId,
			"workflow_id": id,
			"source":      newSource,
			"target":      newTarget,
			"condition":   connection.Condition,
			"created_at":  now,
			"updated_at":  now,
		}).Insert(); err != nil {
			return nil, err
		}

		restoredConnections = append(restoredConnections, &v1.WorkflowConnection{
			Id:        newConnectionId,
			Source:    newSource,
			Target:    newTarget,
			Condition: connection.Condition,
		})
		insertedConnectionCount++
	}
	fmt.Println("[ROLLBACK] connections inserted without workflow update:", insertedConnectionCount)

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	fmt.Println("[ROLLBACK] committed; restored nodes:", insertedNodeCount, "connections:", insertedConnectionCount)

	_, err = db.Model("workflow").Where("id", id).Data(g.Map{
		"version":    version,
		"updated_at": now,
	}).Update()
	if err != nil {
		return nil, err
	}
	fmt.Println("[ROLLBACK] workflow version updated to:", version)

	return s.GetEditorData(ctx, workflowId)
}

func (s *ServiceEditor) createSnapshotVersion(ctx context.Context, tx gdb.TX, workflowId int64, version int, nodes []*v1.WorkflowNodeItem, connections []*v1.WorkflowConnection, now int64) error {
	snapshotJSON, err := json.Marshal(editorSnapshot{Nodes: nodes, Connections: connections})
	if err != nil {
		return err
	}
	_, err = tx.Model("workflow_version").Data(g.Map{
		"workflow_id": workflowId,
		"version":     version,
		"content":     string(snapshotJSON),
		"snapshot":    string(snapshotJSON),
		"created_at":  now,
	}).Insert()
	return err
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
	condition TEXT NOT NULL DEFAULT '',
	created_at BIGINT NOT NULL DEFAULT 0,
	updated_at BIGINT NOT NULL DEFAULT 0
)`)
	if err != nil {
		return err
	}

	if _, err = db.Exec(ctx, `ALTER TABLE workflow_connection ADD COLUMN IF NOT EXISTS condition TEXT NOT NULL DEFAULT ''`); err != nil {
		return err
	}
	if _, err = db.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_workflow_node_workflow_id ON workflow_node(workflow_id)`); err != nil {
		return err
	}
	if _, err = db.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_workflow_connection_workflow_id ON workflow_connection(workflow_id)`); err != nil {
		return err
	}
	if _, err = db.Exec(ctx, `
CREATE TABLE IF NOT EXISTS workflow_version (
	id SERIAL PRIMARY KEY,
	workflow_id INTEGER NOT NULL REFERENCES workflow(id) ON DELETE CASCADE,
	version INTEGER NOT NULL,
	content TEXT NOT NULL DEFAULT '{}',
	snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
	created_at BIGINT NOT NULL DEFAULT 0
)`); err != nil {
		return err
	}
	if _, err = db.Exec(ctx, `ALTER TABLE workflow_version ADD COLUMN IF NOT EXISTS snapshot JSONB NOT NULL DEFAULT '{}'::jsonb`); err != nil {
		return err
	}
	if _, err = db.Exec(ctx, `ALTER TABLE workflow_version ADD COLUMN IF NOT EXISTS content TEXT NOT NULL DEFAULT '{}'`); err != nil {
		return err
	}
	if _, err = db.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_workflow_version_workflow_version ON workflow_version(workflow_id, version)`); err != nil {
		return err
	}
	if _, err = db.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_workflow_version_workflow_id_version ON workflow_version(workflow_id, version)`); err != nil {
		return err
	}
	if _, err = db.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_workflow_version_workflow_version_desc ON workflow_version(workflow_id, version DESC)`); err != nil {
		return err
	}

	return nil
}
