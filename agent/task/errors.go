package task

import "errors"

// 预定义错误
var (
	ErrQueueClosed   = errors.New("task queue is closed")
	ErrTaskNotFound  = errors.New("task not found")
	ErrAgentNotFound = errors.New("agent not found")
	ErrInvalidConfig = errors.New("invalid configuration")
)
