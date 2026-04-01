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
	response := ts.taskChat.TaskChat(ctx, []domain.Message{
		{Role: "user", Content: prompt},
	})

	logger.Info("任务总结生成完成")
	return response, nil
}

// buildSummaryPrompt 构建总结提示词
func (ts *TaskSummarizer) buildSummaryPrompt(originalTask *domain.Task, summary *AggregationSummary) string {
	prompt := fmt.Sprintf(`请为以下任务生成最终总结报告：

## 原始任务
%s

## 执行概览
- 总任务数: %d
- 成功完成: %d
- 失败: %d

## 子任务详情：
`, originalTask.Description, summary.Total, summary.Completed, summary.Failed)

	for i, result := range summary.Results {
		prompt += fmt.Sprintf("\n### 子任务 %d: %s\n", i+1, result.Description)
		prompt += fmt.Sprintf("- 状态: %s\n", result.Status)

		if result.Status == "completed" {
			prompt += fmt.Sprintf("- 输出结果:\n%s\n", result.Output)
		} else {
			prompt += fmt.Sprintf("- 错误信息: %s\n", result.Error)
		}
	}

	prompt += `
请生成一份完整的总结报告，包括：
1. 任务整体完成情况
2. 每个子任务的主要成果
3. 失败任务的影响分析
4. 最终结论和建议

请以清晰、专业的格式呈现。`

	return prompt
}
