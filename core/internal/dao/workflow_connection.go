package dao

import (
	"github.com/gogf/gf/v2/frame/g"
	"billionmail-core/internal/model/entity"
)

var WorkflowConnection = workflowConnectionDao{}

type workflowConnectionDao struct {
	*g.Model
}

func init() {
	WorkflowConnection = workflowConnectionDao{
		g.Model(entity.WorkflowConnection{}).Safe(),
	}
}