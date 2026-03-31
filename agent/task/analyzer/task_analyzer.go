// analyzer/task_analyzer.go
package analyzer

import (
	"encoding/json"
	"fmt"
	"orange-agent/domain"
)

type TaskAnalyzer struct {
	llmClient LLMClient
}

type LLMClient interface {
	Chat(messages []domain.Message) (string, error)
}

func NewTaskAnalyzer(llmClient LLMClient) *TaskAnalyzer {
	return &TaskAnalyzer{
		llmClient: llmClient,
	}
}

// AnalyzeAndSplit 分析并拆分任务
func (ta *TaskAnalyzer) AnalyzeAndSplit(taskDescription string) ([]*domain.SubTask, error) {
	// 第一步：分析任务
	analysisPrompt := fmt.Sprintf(`Analyze the following task and provide a breakdown:

Task: %s

Please output in JSON format:
{
    "task_type": "type of task",
    "complexity": "low/medium/high",
    "estimated_subtasks": number,
    "dependencies": ["list of dependency types"]
}

Analysis:`, taskDescription)

	analysis, err := ta.llmClient.Chat([]domain.Message{
		{Role: "system", Content: "You are a task analysis expert."},
		{Role: "human", Content: analysisPrompt},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to analyze task: %w", err)
	}

	// 第二步：拆分任务
	splitPrompt := fmt.Sprintf(`Based on the analysis:%s, split the following task into smaller, independent subtasks:

Original Task: %s

Requirements for each subtask:
1. Each subtask must be independent and self-contained
2. Each subtask should have clear inputs and expected outputs
3. Subtasks should be executable in parallel when possible
4. Total subtasks should be between 2-8

Output as JSON array:
[
    {
        "description": "detailed description",
        "required_inputs": ["input1", "input2"],
        "expected_output": "description of output",
        "system_prompt": "custom system prompt for this subtask"
    }
]

Subtasks:`, analysis, taskDescription)

	response, err := ta.llmClient.Chat([]domain.Message{
		{Role: "system", Content: "You are a task decomposition expert."},
		{Role: "human", Content: splitPrompt},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to split task: %w", err)
	}

	var subtasksData []struct {
		Description    string   `json:"description"`
		RequiredInputs []string `json:"required_inputs"`
		ExpectedOutput string   `json:"expected_output"`
		SystemPrompt   string   `json:"system_prompt"`
	}

	if err := json.Unmarshal([]byte(response), &subtasksData); err != nil {
		return nil, fmt.Errorf("failed to parse subtasks: %w", err)
	}

	// 构建SubTask对象
	subtasks := make([]*domain.SubTask, 0, len(subtasksData))
	for _, data := range subtasksData {
		subtask := &domain.SubTask{
			Description: data.Description,
			Status:      domain.StatusPending,
			Context:     nil, // 稍后创建
			Input:       make(map[string]interface{}),
			Output:      "",
		}

		// 设置默认系统提示
		if data.SystemPrompt == "" {
			data.SystemPrompt = "You are a helpful assistant focused on completing this specific subtask."
		}

		subtasks = append(subtasks, subtask)
	}

	return subtasks, nil
}
