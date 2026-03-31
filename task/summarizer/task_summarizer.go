// summarizer/task_summarizer.go
package summarizer

import (
	"fmt"
	"orange-agent/domain"
	"strings"
)

type TaskSummarizer struct {
	llmClient LLMClient
}

type LLMClient interface {
	Chat(messages []domain.Message) (string, error)
}

func NewTaskSummarizer(llmClient LLMClient) *TaskSummarizer {
	return &TaskSummarizer{
		llmClient: llmClient,
	}
}

// Summarize 总结所有子任务结果
func (ts *TaskSummarizer) Summarize(originalTask string, subtasks []*domain.SubTask, results []*domain.TaskResult) (string, error) {
	// 构建任务执行摘要
	var summary strings.Builder

	summary.WriteString(fmt.Sprintf("Original Task: %s\n\n", originalTask))
	summary.WriteString("Subtask Results:\n")

	for i, subtask := range subtasks {
		result := results[i]
		status := "✅"
		if !result.Success {
			status = "❌"
		}

		summary.WriteString(fmt.Sprintf("%s Subtask %d: %s\n", status, i+1, subtask.Description))
		if result.Success {
			summary.WriteString(fmt.Sprintf("   Output: %s\n", truncate(subtask.Output, 200)))
		} else {
			summary.WriteString(fmt.Sprintf("   Error: %s\n", result.Error))
		}
		summary.WriteString(fmt.Sprintf("   Tokens: %d | Time: %dms\n", result.TokenUsed, result.ExecutionMs))
	}

	// 调用LLM生成最终总结
	finalPrompt := fmt.Sprintf(`Based on the following task and its execution results, provide a comprehensive summary:

%s

Please provide:
1. Overall completion status
2. Key findings and outputs
3. Any issues encountered
4. Final conclusion

Summary:`, summary.String())

	finalSummary, err := ts.llmClient.Chat([]domain.Message{
		{Role: "system", Content: "You are a summary expert. Provide clear, concise summaries of task executions."},
		{Role: "user", Content: finalPrompt},
	})

	if err != nil {
		return summary.String(), err
	}

	return finalSummary, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
