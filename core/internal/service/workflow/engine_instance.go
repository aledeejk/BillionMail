package workflow

import (
	"billionmail-core/internal/service/workflow/queue"
)

var globalEngine *ExecutionEngine

func InitEngine(q *queue.RedisQueue) {
	globalEngine = NewExecutionEngine(q)
}

func GetExecutionEngine() *ExecutionEngine {
	if globalEngine == nil {
		globalEngine = NewExecutionEngine(nil)
	}
	return globalEngine
}
