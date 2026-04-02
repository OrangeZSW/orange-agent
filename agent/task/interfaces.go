package task

import (
	"context"

	"orange-agent/domain"
)

// TaskChat 任务聊天接口
type TaskChat interface {
	Chat(ctx context.Context, messages []domain.Message) string
}

// TaskExecutorInterface 子任务执行器接口
type TaskExecutorInterface interface {
	ExecuteSubTask(ctx context.Context, subTask *domain.SubTask) error
}

// TaskSplitterInterface 任务拆分器接口
type TaskSplitterInterface interface {
	Split(ctx context.Context, task *domain.Task, analysis *AnalysisResult) ([]*domain.SubTask, error)
}

// TaskAnalyzerInterface 任务分析器接口
type TaskAnalyzerInterface interface {
	Analyze(ctx context.Context, taskDescription string) (*AnalysisResult, error)
}

// DAGEngineInterface 依赖图执行引擎接口
type DAGEngineInterface interface {
	ExecuteDAG(ctx context.Context, task *domain.Task) (string, error)
	BuildDAG(subTasks []*domain.SubTask) (*domain.DependencyGraph, error)
}

// ResultAggregatorInterface 结果聚合器接口
type ResultAggregatorInterface interface {
	AddResult(subTask *domain.SubTask)
	GetSummary() *AggregationSummary
	GetProgress() float64
	GetSuccessCount() int
	GetFailedCount() int
}

// ContextManagerInterface 上下文管理器接口
type ContextManagerInterface interface {
	CreateTaskContext(taskID uint, systemPrompt string) *domain.TaskContext
	AddMessage(taskID uint, role, content string, tokenCount int)
	GetContext(taskID uint) *domain.TaskContext
	CompressContext(taskID uint, maxTokens int) error
}

// TaskSummarizerInterface 任务总结器接口
type TaskSummarizerInterface interface {
	Summarize(ctx context.Context, task *domain.Task, summary *AggregationSummary) (string, error)
}