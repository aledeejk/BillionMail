package dao

import (
	"github.com/gogf/gf/v2/frame/g"
	"billionmail-core/internal/model/entity"
)

var Workflow = workflowDao{}

type workflowDao struct {
	*g.Model
}

func init() {
	Workflow = workflowDao{
		g.Model(entity.Workflow{}).Safe(),
	}
}