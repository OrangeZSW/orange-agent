package task

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"orange-agent/domain"
	"orange-agent/utils/logger"
)

// TaskOrchestrator 任务编排器，协调整个流程
type TaskOrchestrator struct {
	contextManager   ContextManagerInterface
	taskAnalyzer     TaskAnalyzerInterface
	taskSplitter     TaskSplitterInterface
	resultAggregator ResultAggregatorInterface
	taskSummarizer   TaskSummarizerInterface
	dagEngine        DAGEngineInterface
	taskExecutor     TaskExecutorInterface
	taskChat         TaskChat
	config           *OrchestratorConfig
}

// OrchestratorConfig 编排器配置
type OrchestratorConfig struct {
	WorkerCount     int
	QueueBufferSize int
	UseDAGEngine    bool // 是否使用DAG引擎
}

// DefaultOrchestratorConfig 默认配置
func DefaultOrchestratorConfig() *OrchestratorConfig {
	return &OrchestratorConfig{
		WorkerCount:     3,
		QueueBufferSize: 100,
		UseDAGEngine:    true, // 默认使用DAG引擎
	}
}

// OrchestratorOption 构造函数选项
type OrchestratorOption func(*TaskOrchestrator)

// WithContextManager 自定义上下文管理器
func WithContextManager(cm ContextManagerInterface) OrchestratorOption {
	return func(to *TaskOrchestrator) {
		to.contextManager = cm
	}
}

// WithTaskAnalyzer 自定义任务分析器
func WithTaskAnalyzer(ta TaskAnalyzerInterface) OrchestratorOption {
	return func(to *TaskOrchestrator) {
		to.taskAnalyzer = ta
	}
}

// WithTaskSplitter 自定义任务拆分器
func WithTaskSplitter(ts TaskSplitterInterface) OrchestratorOption {
	return func(to *TaskOrchestrator) {
		to.taskSplitter = ts
	}
}

// WithTaskSummarizer 自定义任务总结器
func WithTaskSummarizer(ts TaskSummarizerInterface) OrchestratorOption {
	return func(to *TaskOrchestrator) {
		to.taskSummarizer = ts
	}
}

// WithDAGEngine 自定义DAG引擎
func WithDAGEngine(de DAGEngineInterface) OrchestratorOption {
	return func(to *TaskOrchestrator) {
		to.dagEngine = de
	}
}

// WithTaskExecutor 自定义子任务执行器
func WithTaskExecutor(te TaskExecutorInterface) OrchestratorOption {
	return func(to *TaskOrchestrator) {
		to.taskExecutor = te
	}
}

// NewTaskOrchestrator 创建新的任务编排器（依赖注入模式）
func NewTaskOrchestrator(config *OrchestratorConfig, taskChat TaskChat, opts ...OrchestratorOption) *TaskOrchestrator {
	if config == nil {
		config = DefaultOrchestratorConfig()
	}

	to := &TaskOrchestrator{
		taskChat: taskChat,
		config:   config,
	}

	// 初始化默认组件
	to.contextManager = NewContextManager()
	to.taskAnalyzer = NewTaskAnalyzer(taskChat)
	to.taskSplitter = NewTaskSplitter(taskChat)
	to.taskSummarizer = NewTaskSummarizer(taskChat)
	to.dagEngine = NewDAGEngine(taskChat, config.WorkerCount)
	to.taskExecutor = NewTaskExecutor(to.contextManager, taskChat)

	// 应用自定义选项
	for _, opt := range opts {
		opt(to)
	}

	return to
}

// Execute 执行整个任务流程
func (to *TaskOrchestrator) Execute(ctx context.Context, task *domain.Task) (string, error) {
	logger.Info("开始执行任务: %s", task.Description)

	// 1. 分析任务
	analysis, err := to.analyzeTask(ctx, task.Description)
	if err != nil {
		return "", fmt.Errorf("任务分析失败: %w", err)
	}

	// 2. 拆分任务
	subTasks, err := to.splitTask(ctx, task, analysis)
	if err != nil {
		return "", fmt.Errorf("任务拆分失败: %w", err)
	}

	task.Subtasks = subTasks

	// 3. 选择执行引擎并执行
	var summary *AggregationSummary
	if to.shouldUseDAGEngine(subTasks) {
		summary, err = to.executeWithDAGEngine(ctx, task)
	} else {
		summary, err = to.executeSequential(ctx, task, subTasks)
	}
	if err != nil {
		return "", err
	}

	// 4. 生成最终总结
	finalSummary, err := to.generateFinalSummary(ctx, task, summary)
	if err != nil {
		return "", fmt.Errorf("生成总结失败: %w", err)
	}

	task.Result = finalSummary
	task.Status = domain.StatusCompleted
	logger.Info("任务执行流程全部完成")

	return finalSummary, nil
}

// analyzeTask 封装任务分析逻辑
func (to *TaskOrchestrator) analyzeTask(ctx context.Context, description string) (*AnalysisResult, error) {
	logger.Info("开始分析任务...")
	analysis, err := to.taskAnalyzer.Analyze(ctx, description)
	if err != nil {
		return nil, err
	}
	logger.Info("任务分析完成")
	return analysis, nil
}

// splitTask 封装任务拆分逻辑
func (to *TaskOrchestrator) splitTask(ctx context.Context, task *domain.Task, analysis *AnalysisResult) ([]*domain.SubTask, error) {
	logger.Info("开始拆分任务...")
	subTasks, err := to.taskSplitter.Split(ctx, task, analysis)
	if err != nil {
		return nil, err
	}
	if len(subTasks) == 0 {
		return nil, errors.New("没有生成任何子任务")
	}
	logger.Info("拆分任务成功，共生成 %d 个子任务", len(subTasks))
	return subTasks, nil
}

// shouldUseDAGEngine 判断是否应该使用DAG引擎
func (to *TaskOrchestrator) shouldUseDAGEngine(subTasks []*domain.SubTask) bool {
	// 如果配置强制不使用DAG引擎
	if !to.config.UseDAGEngine {
		return false
	}

	// 检查是否有显式的依赖关系
	hasExplicitDependencies := false
	hasComplexDependencies := false

	for _, task := range subTasks {
		if len(task.Dependencies) > 0 {
			hasExplicitDependencies = true
		}
		if len(task.Dependencies) > 1 {
			hasComplexDependencies = true
		}
	}

	// 如果有复杂的依赖关系，使用DAG引擎
	if hasComplexDependencies {
		return true
	}

	// 如果有显式依赖关系且不是简单的线性顺序，使用DAG引擎
	if hasExplicitDependencies {
		// 检查是否所有任务都标记为可并行
		allParallel := true
		for _, task := range subTasks {
			if !task.CanParallel {
				allParallel = false
				break
			}
		}
		return !allParallel // 如果有依赖但不可完全并行，使用DAG
	}

	// 默认使用顺序引擎
	return false
}

// executeWithDAGEngine 使用DAG引擎执行任务
func (to *TaskOrchestrator) executeWithDAGEngine(ctx context.Context, task *domain.Task) (*AggregationSummary, error) {
	logger.Info("选择DAG引擎执行任务")
	_, err := to.dagEngine.ExecuteDAG(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("DAG执行失败: %w", err)
	}
	return to.buildSummaryFromSubTasks(task.Subtasks), nil
}

// executeSequential 顺序执行子任务
func (to *TaskOrchestrator) executeSequential(ctx context.Context, task *domain.Task, subTasks []*domain.SubTask) (*AggregationSummary, error) {
	logger.Info("选择顺序引擎执行任务")
	// 按执行顺序排序
	sort.Slice(subTasks, func(i, j int) bool {
		return subTasks[i].ExecutionOrder < subTasks[j].ExecutionOrder
	})

	// 创建结果聚合器
	aggregator := NewResultAggregator(subTasks)

	// 顺序执行子任务，前一个的结果传给后一个
	var previousResult string
	for i, subTask := range subTasks {
		// 如果不是第一个任务，将前一个任务的结果作为输入
		if i > 0 && previousResult != "" {
			if subTask.Input == nil {
				subTask.Input = make(map[string]interface{})
			}
			subTask.Input["previous_result"] = previousResult
			subTask.Input["previous_subtask_index"] = i - 1
		}

		// 执行当前子任务
		logger.Info("开始执行任务%d (顺序:%d): %s", i+1, subTask.ExecutionOrder, subTask.Description)
		if err := to.taskExecutor.ExecuteSubTask(ctx, subTask); err != nil {
			subTask.Status = domain.StatusFailed
			subTask.Error = err.Error()
			logger.Error("任务%d执行失败: %s，终止后续任务", i+1, err.Error())
			aggregator.AddResult(subTask)
			break
		}

		// 收集结果
		aggregator.AddResult(subTask)
		logger.Info("任务%d执行结果: 成功", i+1)

		// 保存当前任务结果作为下一个任务的输入
		previousResult = subTask.Output
	}

	// 获取聚合摘要
	return aggregator.GetSummary(), nil
}

// buildSummaryFromSubTasks 从子任务构建摘要
func (to *TaskOrchestrator) buildSummaryFromSubTasks(subTasks []*domain.SubTask) *AggregationSummary {
	// 创建结果聚合器
	resultAggregator := NewResultAggregator(subTasks)

	// 添加所有子任务结果
	for _, subTask := range subTasks {
		resultAggregator.AddResult(subTask)
	}

	// 获取聚合摘要
	return resultAggregator.GetSummary()
}

// generateFinalSummary 生成最终任务总结
func (to *TaskOrchestrator) generateFinalSummary(ctx context.Context, task *domain.Task, summary *AggregationSummary) (string, error) {
	logger.Info("开始生成任务总结...")
	finalSummary, err := to.taskSummarizer.Summarize(ctx, task, summary)
	if err != nil {
		return "", err
	}
	logger.Info("任务总结生成完成")
	return finalSummary, nil
}

// GetResultAggregator 获取结果聚合器
func (to *TaskOrchestrator) GetResultAggregator() ResultAggregatorInterface {
	return to.resultAggregator
}

// GetDAGEngine 获取DAG引擎
func (to *TaskOrchestrator) GetDAGEngine() DAGEngineInterface {
	return to.dagEngine
}
