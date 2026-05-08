package workflow

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"

	v1 "github.com/BillionMail/BillionMail/core/api/workflow/v1"
	"github.com/BillionMail/BillionMail/core/internal/dao"
	"github.com/BillionMail/BillionMail/core/internal/model/entity"
)

type ServiceEditor struct{}

func (s *ServiceEditor) GetEditorData(ctx context.Context, workflowId int64) (*v1.GetWorkflowEditorRes, error) {
	// Get workflow
	workflow, err := dao.Workflow.Ctx(ctx).Where("id", workflowId).One()
	if err != nil {
		return nil, err
	}
	if workflow.IsEmpty() {
		return nil, gerror.New("workflow not found")
	}

	// Get nodes
	nodes := make([]*v1.WorkflowNodeItem, 0)
	err = dao.WorkflowNode.Ctx(ctx).Where("workflow_id", workflowId).Scan(&nodes)
	if err != nil {
		return nil, err
	}

	// Get connections
	connections := make([]*v1.WorkflowConnection, 0)
	err = dao.WorkflowConnection.Ctx(ctx).Where("workflow_id", workflowId).Scan(&connections)
	if err != nil {
		return nil, err
	}

	result := &v1.GetWorkflowEditorRes{
		Workflow: &v1.WorkflowItem{
			Id:          workflow["id"].Int64(),
			Name:        workflow["name"].String(),
			Description: workflow["description"].String(),
			IsActive:    workflow["is_active"].Bool(),
			Version:     workflow["version"].Int(),
			CreatedAt:   workflow["created_at"].String(),
			UpdatedAt:   workflow["updated_at"].String(),
		},
		Nodes:       nodes,
		Connections: connections,
	}

	return result, nil
}

func (s *ServiceEditor) UpdateEditorData(ctx context.Context, workflowId int64, nodes []*v1.WorkflowNodeItem, connections []*v1.WorkflowConnection) error {
	return dao.WorkflowNode.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Delete existing nodes and connections
		_, err := dao.WorkflowNode.Ctx(ctx).Where("workflow_id", workflowId).Delete()
		if err != nil {
			return err
		}

		_, err = dao.WorkflowConnection.Ctx(ctx).Where("workflow_id", workflowId).Delete()
		if err != nil {
			return err
		}

		// Insert new nodes
		for _, node := range nodes {
			_, err = dao.WorkflowNode.Ctx(ctx).Data(gdb.Map{
				"workflow_id": workflowId,
				"node_id":     node.Id,
				"type":        node.Type,
				"config":      node.Config,
				"position_x":  node.PositionX,
				"position_y":  node.PositionY,
			}).Insert()
			if err != nil {
				return err
			}
		}

		// Insert new connections
		for _, conn := range connections {
			_, err = dao.WorkflowConnection.Ctx(ctx).Data(gdb.Map{
				"workflow_id": workflowId,
				"connection_id": conn.Id,
				"source_node_id": conn.Source,
				"target_node_id": conn.Target,
			}).Insert()
			if err != nil {
				return err
			}
		}

		// Update workflow updated_at
		_, err = dao.Workflow.Ctx(ctx).Where("id", workflowId).Data(gdb.Map{
			"updated_at": gdb.Raw("NOW()"),
		}).Update()
		if err != nil {
			return err
		}

		return nil
	})
}