package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"orange-agent/domain"
	"orange-agent/utils/logger"
)

// TaskAnalyzer 分析用户输入的总任务
type TaskAnalyzer struct {
	TaskChat TaskChat
}

// NewTaskAnalyzer 创建新的任务分析器
func NewTaskAnalyzer(taskChat TaskChat) *TaskAnalyzer {
	return &TaskAnalyzer{
		TaskChat: taskChat,
	}
}

// AnalysisResult 任务分析结果
type AnalysisResult struct {
	TaskType          string   `json:"task_type"`
	Complexity        string   `json:"complexity"` // low, medium, high
	EstimatedSubtasks int      `json:"estimated_subtasks"`
	KeyObjectives     []string `json:"key_objectives"`
	Constraints       []string `json:"constraints"`
	RecommendEngine   string   `json:"recommend_engine"` // sequential, dag, parallel
	EstimatedTime     int      `json:"estimated_time"`   // 预估执行时间（分钟）
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
  "constraints": ["约束条件1", "约束条件2"],
  "recommend_engine": "推荐执行引擎（sequential-顺序/dag-依赖图/parallel-并行）",
  "estimated_time": 预估执行时间（分钟）
}

分析指导原则：
1. 任务复杂度：根据任务规模、技术难度、协调需求等评估
2. 子任务数量：预估需要拆分成多少个独立的子任务
3. 执行引擎推荐：
   - sequential: 简单的线性任务，子任务之间有强依赖
   - dag: 复杂的依赖关系，部分任务可并行执行
   - parallel: 高度独立的子任务，可以完全并行执行

只返回JSON，不要其他内容。`, taskDescription)

	response := ta.TaskChat.Chat(ctx, []domain.Message{
		{Role: "user", Content: prompt},
	})

	// 解析JSON响应
	var result AnalysisResult
	// 清理响应内容，只保留JSON部分
	jsonStr := extractJSONFromResponse(response)
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("解析分析结果失败: %w", err)
	}

	// 验证推荐引擎值
	validEngines := map[string]bool{
		"sequential": true,
		"dag":        true,
		"parallel":   true,
	}
	if !validEngines[result.RecommendEngine] {
		result.RecommendEngine = "sequential" // 默认值
	}

	logger.Info("任务分析完成: %+v", result)
	return &result, nil
}

// extractJSONFromResponse 从字符串中提取JSON部分
func extractJSONFromResponse(content string) string {
	// 查找第一个 { 和最后一个 }
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start == -1 || end == -1 || start >= end {
		return content
	}
	return content[start : end+1]
}