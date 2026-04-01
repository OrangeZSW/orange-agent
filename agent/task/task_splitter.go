package task

import (
	"context"
	"encoding/json"
	"fmt"

	"orange-agent/domain"
	"orange-agent/utils/logger"
)

// TaskSplitter 将总任务拆分为多个子任务
type TaskSplitter struct {
	taskChat TaskChat
}

// NewTaskSplitter 创建新的任务分割器
func NewTaskSplitter(taskChat TaskChat) *TaskSplitter {
	return &TaskSplitter{
		taskChat: taskChat,
	}
}

// Split 将总任务拆分为子任务
func (ts *TaskSplitter) Split(ctx context.Context, task *domain.Task, analysis *AnalysisResult) ([]*domain.SubTask, error) {
	logger.Info("开始拆分任务: %s", task.Description)

	prompt := fmt.Sprintf(`请将以下总任务拆分为具体的子任务，并以JSON格式返回：

总任务：%s
任务分析：
- 类型: %s
- 复杂度: %s
- 预估子任务数: %d
- 关键目标: %v
- 约束条件: %v

请返回以下格式的JSON：
{
  "subtasks": [
    {
      "description": "子任务描述",
      "input": {
        "key": "value"
      }
    }
  ]
}

子任务应该：
1. 每个子任务相对独立，可并行执行
2. 有明确的输入输出
3. 任务描述要具体可执行
4. 输入部分要包含该子任务需要的所有信息

只返回JSON，不要其他内容。`,
		task.Description,
		analysis.TaskType,
		analysis.Complexity,
		analysis.EstimatedSubtasks,
		analysis.KeyObjectives,
		analysis.Constraints)

	// 获取默认agent进行任务拆分
	response := ts.taskChat.Chat(ctx, []domain.Message{
		{
			Role:    "user",
			Content: prompt,
		},
	})

	// 解析JSON响应
	var result struct {
		Subtasks []struct {
			Description string                 `json:"description"`
			Input       map[string]interface{} `json:"input"`
		} `json:"subtasks"`
	}

	// 清理响应内容，只保留JSON部分
	jsonStr := extractJSON(response)
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("解析拆分结果失败: %w", err)
	}

	// 创建SubTask对象
	var subTasks []*domain.SubTask
	for _, st := range result.Subtasks {
		subTask := &domain.SubTask{
			Description: st.Description,
			Status:      domain.StatusPending,
			Input:       st.Input,
			TaskID:      task.ID,
		}
		subTasks = append(subTasks, subTask)
	}

	logger.Info("任务拆分完成，共拆分为 %d 个子任务", len(subTasks))
	return subTasks, nil
}
