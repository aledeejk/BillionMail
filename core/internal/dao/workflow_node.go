package dao

import (
	"github.com/gogf/gf/v2/frame/g"
	"billionmail-core/internal/model/entity"
)

var WorkflowNode = workflowNodeDao{}

type workflowNodeDao struct {
	*g.Model
}

func init() {
	WorkflowNode = workflowNodeDao{
		g.Model(entity.WorkflowNode{}).Safe(),
	}
}