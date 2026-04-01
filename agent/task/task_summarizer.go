package task

import (
	"context"
	"fmt"

	"orange-agent/domain"
	"orange-agent/utils/logger"
)

// TaskSummarizer 总结聚合结果，生成最终输出
type TaskSummarizer struct {
	taskChat TaskChat
}

// NewTaskSummarizer 创建新的任务总结器
func NewTaskSummarizer(taskChat TaskChat) *TaskSummarizer {
	return &TaskSummarizer{
		taskChat: taskChat,
	}
}

// Summarize 生成最终总结
func (ts *TaskSummarizer) Summarize(ctx context.Context, originalTask *domain.Task, summary *AggregationSummary) (string, error) {
	logger.Info("开始生成任务总结")

	// 构建提示词
	prompt := ts.buildSummaryPrompt(originalTask, summary)

	// 调用agent生成总结
	response := ts.taskChat.Chat(ctx, []domain.Message{
		{Role: "user", Content: prompt},
	})

	logger.Info("任务总结生成完成")
	return response, nil
}

// buildSummaryPrompt 构建总结提示词
func (ts *TaskSummarizer) buildSummaryPrompt(originalTask *domain.Task, summary *AggregationSummary) string {
	prompt := fmt.Sprintf(`请为以下任务生成最终总结报告，按照以下格式组织：

## 任务执行流程

开始分析任务 -》 拆分任务成功，%d 个任务：
`, summary.Total)

	for i, result := range summary.Results {
		prompt += fmt.Sprintf("任务%d: %s\n", i+1, result.Description)
	}

	for i, result := range summary.Results {
		prompt += fmt.Sprintf("\n-》任务%d执行：%s\n", i+1, result.Description)
		prompt += fmt.Sprintf("-》任务%d执行结果：", i+1)
		if result.Status == "completed" {
			prompt += "成功\n"
			prompt += fmt.Sprintf("%s\n", result.Output)
		} else {
			prompt += fmt.Sprintf("失败 - %s\n", result.Error)
		}
	}

	prompt += "\n-》任务总结"

	prompt += fmt.Sprintf(`

## 原始任务
%s

## 执行概览
- 总任务数: %d
- 成功完成: %d
- 失败: %d

请按照上面的"任务执行流程"格式，生成一份完整的总结报告，保持清晰的流程感。`, originalTask.Description, summary.Total, summary.Completed, summary.Failed)

	return prompt
}
