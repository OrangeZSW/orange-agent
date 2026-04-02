package task

import (
	"context"
	"fmt"

	"orange-agent/domain"
	"orange-agent/utils/logger"
)

// TaskExecutor 子任务执行器实现
type TaskExecutor struct {
	contextManager ContextManagerInterface
	taskChat       TaskChat
}

// NewTaskExecutor 创建新的子任务执行器
func NewTaskExecutor(contextManager ContextManagerInterface, taskChat TaskChat) *TaskExecutor {
	return &TaskExecutor{
		contextManager: contextManager,
		taskChat:       taskChat,
	}
}

// ExecuteSubTask 执行单个子任务
func (te *TaskExecutor) ExecuteSubTask(ctx context.Context, subTask *domain.SubTask) error {
	// 创建独立的任务上下文
	taskCtx := te.contextManager.CreateTaskContext(
		subTask.ID,
		"You are a helpful assistant that executes tasks efficiently.",
	)
	subTask.Context = taskCtx

	// 构建任务提示
	prompt := te.buildTaskPrompt(subTask)

	// 添加用户消息到上下文
	te.contextManager.AddMessage(subTask.ID, "user", prompt, len(prompt))

	// 执行任务
	messages := []domain.Message{
		{Role: "system", Content: taskCtx.SystemPrompt},
		{Role: "user", Content: prompt},
	}

	response := te.taskChat.Chat(ctx, messages)

	// 添加助手响应到上下文
	te.contextManager.AddMessage(subTask.ID, "assistant", response, len(response))

	// 更新子任务状态
	subTask.Status = domain.StatusCompleted
	subTask.Output = response

	logger.Info("子任务执行成功: %s", subTask.Description)
	return nil
}

// buildTaskPrompt 构建任务提示
func (te *TaskExecutor) buildTaskPrompt(subTask *domain.SubTask) string {
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
