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
	contextManager   *ContextManager
	taskAnalyzer     *TaskAnalyzer
	taskSplitter     *TaskSplitter
	resultAggregator *ResultAggregator
	taskSummarizer   *TaskSummarizer
	agentManager     interfaces.AgentManager
	taskChat         TaskChat
}

// OrchestratorConfig 编排器配置
type OrchestratorConfig struct {
	WorkerCount     int
	QueueBufferSize int
}

// DefaultOrchestratorConfig 默认配置
func DefaultOrchestratorConfig() *OrchestratorConfig {
	return &OrchestratorConfig{
		WorkerCount:     3,
		QueueBufferSize: 100,
	}
}

// NewTaskOrchestrator 创建新的任务编排器
func NewTaskOrchestrator(config *OrchestratorConfig, taskChat TaskChat) *TaskOrchestrator {
	if config == nil {
		config = DefaultOrchestratorConfig()
	}

	contextManager := NewContextManager()

	return &TaskOrchestrator{
		contextManager: contextManager,
		taskAnalyzer:   NewTaskAnalyzer(taskChat),
		taskSplitter:   NewTaskSplitter(taskChat),
		taskSummarizer: NewTaskSummarizer(taskChat),
		taskChat:       taskChat,
	}
}

// Execute 执行整个任务流程
func (to *TaskOrchestrator) Execute(ctx context.Context, task *domain.Task) (string, error) {
	logger.Info("开始执行任务: %s", task.Description)

	// 1. 分析任务
	logger.Info("开始分析任务...")
	analysis, err := to.taskAnalyzer.Analyze(ctx, task.Description)
	if err != nil {
		return "", fmt.Errorf("任务分析失败: %w", err)
	}
	logger.Info("任务分析完成")

	// 2. 拆分任务
	logger.Info("开始拆分任务...")
	subTasks, err := to.taskSplitter.Split(ctx, task, analysis)
	if err != nil {
		return "", fmt.Errorf("任务拆分失败: %w", err)
	}

	if len(subTasks) == 0 {
		return "", errors.New("没有生成任何子任务")
	}

	task.Subtasks = subTasks
	logger.Info("拆分任务成功，共生成 %d 个任务:", len(subTasks))
	for i, subTask := range subTasks {
		logger.Info("  任务%d: %s", i+1, subTask.Description)
	}

	// 3. 创建结果聚合器
	to.resultAggregator = NewResultAggregator(subTasks)

	// 4. 顺序执行子任务，前一个的结果传给后一个
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
		logger.Info("开始执行任务%d: %s", i+1, subTask.Description)
		to.executeSubTask(ctx, subTask)

		// 收集结果
		to.resultAggregator.AddResult(subTask)

		// 记录任务执行结果
		if subTask.Status == domain.StatusFailed {
			logger.Info("任务%d执行结果: 失败 - %s", i+1, subTask.Error)
			logger.Error("任务%d执行失败，终止后续任务", i+1)
			break
		} else {
			logger.Info("任务%d执行结果: 成功", i+1)
		}

		// 保存当前任务结果作为下一个任务的输入
		previousResult = subTask.Output
	}

	// 5. 获取聚合摘要
	summary := to.resultAggregator.GetSummary()

	// 6. 生成最终总结
	logger.Info("开始生成任务总结...")
	finalResult, err := to.taskSummarizer.Summarize(ctx, task, summary)
	if err != nil {
		return "", fmt.Errorf("生成总结失败: %w", err)
	}

	task.Result = finalResult
	task.Status = domain.StatusCompleted

	logger.Info("任务总结生成完成")
	logger.Info("任务执行流程全部完成")
	return finalResult, nil
}

// executeSubTask 执行单个子任务
func (to *TaskOrchestrator) executeSubTask(ctx context.Context, subTask *domain.SubTask) {
	// 创建独立的任务上下文
	taskCtx := to.contextManager.CreateTaskContext(
		subTask.ID,
		"You are a helpful assistant that executes tasks efficiently.",
	)
	subTask.Context = taskCtx

	// 构建任务提示
	prompt := to.buildTaskPrompt(subTask)

	// 添加用户消息到上下文
	to.contextManager.AddMessage(subTask.ID, "user", prompt, len(prompt))

	// 执行任务
	messages := []domain.Message{
		{Role: "system", Content: taskCtx.SystemPrompt},
		{Role: "user", Content: prompt},
	}

	response := to.taskChat.Chat(ctx, messages)

	// 添加助手响应到上下文
	to.contextManager.AddMessage(subTask.ID, "assistant", response, len(response))

	// 更新子任务状态
	subTask.Status = domain.StatusCompleted
	subTask.Output = response

	logger.Info("子任务执行成功: %s", subTask.Description)
}

// buildTaskPrompt 构建任务提示
func (to *TaskOrchestrator) buildTaskPrompt(subTask *domain.SubTask) string {
	prompt := fmt.Sprintf("请执行以下任务：\n\n任务描述：%s\n\n", subTask.Description)

	if len(subTask.Input) > 0 {
		prompt += "输入信息：\n"
		for key, value := range subTask.Input {
			if key == "previous_result" {
				prompt += fmt.Sprintf("- 前一个子任务的执行结果：\n%s\n", value)
			} else {
				prompt += fmt.Sprintf("- %s: %v\n", key, value)
			}
		}
	}

	prompt += "\n请详细完成该任务，并返回结果。"
	return prompt
}

// GetResultAggregator 获取结果聚合器
func (to *TaskOrchestrator) GetResultAggregator() *ResultAggregator {
	return to.resultAggregator
}
