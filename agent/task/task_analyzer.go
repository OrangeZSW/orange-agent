package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	
	"orange-agent/agent/interfaces"
	"orange-agent/domain"
	"orange-agent/utils/logger"
)

// TaskAnalyzer 分析用户输入的总任务
type TaskAnalyzer struct {
	agentManager interfaces.AgentManager
}

// NewTaskAnalyzer 创建新的任务分析器
func NewTaskAnalyzer(agentManager interfaces.AgentManager) *TaskAnalyzer {
	return &TaskAnalyzer{
		agentManager: agentManager,
	}
}

// AnalysisResult 任务分析结果
type AnalysisResult struct {
	TaskType        string   `json:"task_type"`
	Complexity      string   `json:"complexity"` // low, medium, high
	EstimatedSubtasks int    `json:"estimated_subtasks"`
	KeyObjectives   []string `json:"key_objectives"`
	Constraints     []string `json:"constraints"`
}

// Analyze 分析用户任务
func (ta *TaskAnalyzer) Analyze(ctx context.Context, taskDescription string) (*AnalysisResult, error) {
	logger.Info("开始分析任务: %s", taskDescription)
	
	prompt := fmt.Sprintf(`请分析以下任务，并以JSON格式返回分析结果：

任务描述：%s

请返回以下格式的JSON：
{
  "task_type": "任务类型（如：代码开发、文档撰写、数据分析等）",
  "complexity": "复杂度（low/medium/high）",
  "estimated_subtasks": 预估的子任务数量,
  "key_objectives": ["关键目标1", "关键目标2"],
  "constraints": ["约束条件1", "约束条件2"]
}

只返回JSON，不要其他内容。`, taskDescription)

	// 获取默认agent进行分析
	agent, err := ta.agentManager.GetDefaultAgent()
	if err != nil {
		return nil, fmt.Errorf("获取默认agent失败: %w", err)
	}

	// 调用agent进行分析
	response, err := agent.Chat(ctx, []domain.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		return nil, fmt.Errorf("任务分析失败: %w", err)
	}

	// 解析JSON响应
	var result AnalysisResult
	// 清理响应内容，只保留JSON部分
	jsonStr := extractJSON(response.Content)
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("解析分析结果失败: %w", err)
	}

	logger.Info("任务分析完成: %+v", result)
	return &result, nil
}

// extractJSON 从字符串中提取JSON部分
func extractJSON(content string) string {
	// 查找第一个 { 和最后一个 }
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start == -1 || end == -1 || start >= end {
		return content
	}
	return content[start : end+1]
}
