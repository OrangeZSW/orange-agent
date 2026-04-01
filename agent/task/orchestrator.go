package task

import (
	"context"
	"errors"
	"fmt"
	
	"orange-agent/agent/interfaces"
	"orange-agent/domain"
	"orange-agent/utils/logger"
)

// TaskOrchestrator 任务编排器，协调整个流程
type TaskOrchestrator struct {
	contextManager *ContextManager
	taskAnalyzer   *TaskAnalyzer
	taskSplitter   *TaskSplitter
	taskQueue      *TaskQueue
	workerPool     *WorkerPool
	resultAggregator *ResultAggregator
	taskSummarizer *TaskSummarizer
	agentManager   interfaces.AgentManager
}

// OrchestratorConfig 编排器配置
type OrchestratorConfig struct {
	WorkerCount   int
	QueueBufferSize int
}

// DefaultOrchestratorConfig 默认配置
func DefaultOrchestratorConfig() *OrchestratorConfig {
	return &OrchestratorConfig{
		WorkerCount:    3,
		QueueBufferSize: 100,
	}
}

// NewTaskOrchestrator 创建新的任务编排器
func NewTaskOrchestrator(agentManager interfaces.AgentManager, config *OrchestratorConfig) *TaskOrchestrator {
	if config == nil {
		config = DefaultOrchestratorConfig()
	}
	
	contextManager := NewContextManager()
	taskQueue := NewTaskQueue(config.QueueBufferSize)
	
	return &TaskOrchestrator{
		contextManager: contextManager,
		taskAnalyzer:   NewTaskAnalyzer(agentManager),
		taskSplitter:   NewTaskSplitter(agentManager),
		taskQueue:      taskQueue,
		workerPool:     NewWorkerPool(config.WorkerCount, taskQueue, agentManager, contextManager),
		taskSummarizer: NewTaskSummarizer(agentManager),
		agentManager:   agentManager,
	}
}

// Execute 执行整个任务流程
func (to *TaskOrchestrator) Execute(ctx context.Context, task *domain.Task) (string, error) {
	logger.Info("开始执行任务编排流程")
	
	// 1. 分析任务
	analysis, err := to.taskAnalyzer.Analyze(ctx, task.Description)
	if err != nil {
		return "", fmt.Errorf("任务分析失败: %w", err)
	}
	
	// 2. 拆分任务
	subTasks, err := to.taskSplitter.Split(ctx, task, analysis)
	if err != nil {
		return "", fmt.Errorf("任务拆分失败: %w", err)
	}
	
	if len(subTasks) == 0 {
		return "", errors.New("没有生成任何子任务")
	}
	
	task.Subtasks = subTasks
	
	// 3. 创建结果聚合器
	to.resultAggregator = NewResultAggregator(subTasks)
	
	// 4. 启动工作池
	to.workerPool.Start()
	defer to.workerPool.Stop()
	
	// 5. 将子任务加入队列
	for _, subTask := range subTasks {
		if err := to.taskQueue.Enqueue(subTask); err != nil {
			logger.Error("加入任务队列失败: %v", err)
		}
	}
	
	// 6. 关闭队列（不再接受新任务）
	to.taskQueue.Close()
	
	// 7. 收集结果
	resultChan := to.workerPool.GetResultChan()
	for subTask := range resultChan {
		to.resultAggregator.AddResult(subTask)
		
		// 检查是否所有任务都已完成
		if to.resultAggregator.IsComplete() {
			break
		}
	}
	
	// 8. 获取聚合摘要
	summary := to.resultAggregator.GetSummary()
	
	// 9. 生成最终总结
	finalResult, err := to.taskSummarizer.Summarize(ctx, task, summary)
	if err != nil {
		return "", fmt.Errorf("生成总结失败: %w", err)
	}
	
	task.Result = finalResult
	task.Status = domain.StatusCompleted
	
	logger.Info("任务编排流程执行完成")
	return finalResult, nil
}

// GetResultAggregator 获取结果聚合器
func (to *TaskOrchestrator) GetResultAggregator() *ResultAggregator {
	return to.resultAggregator
}
