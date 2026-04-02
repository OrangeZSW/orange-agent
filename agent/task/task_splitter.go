package task

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"orange-agent/domain"
	"orange-agent/utils/logger"
	"gorm.io/gorm"
)

// TaskSplitter 将总任务拆分为多个子任务
type TaskSplitter struct {
	taskChat TaskChat
	log      logger.Logger
}

// NewTaskSplitter 创建新的任务分割器
func NewTaskSplitter(taskChat TaskChat) *TaskSplitter {
	return &TaskSplitter{
		taskChat: taskChat,
		log:      *logger.GetLogger(),
	}
}

// Split 将总任务拆分为子任务
func (ts *TaskSplitter) Split(ctx context.Context, task *domain.Task, analysis *AnalysisResult) ([]*domain.SubTask, error) {
	ts.log.Info("开始拆分任务: %s", task.Description)

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
      },
      "dependencies": ["task_index_1", "task_index_2"],  // 可选，依赖的其他任务索引
      "execution_order": 0,  // 执行顺序，从0开始
      "can_parallel": false  // 是否可并行执行
    }
  ]
}

拆分指导原则：
1. 考虑任务之间的依赖关系，明确标注dependencies字段
2. 子任务应该按顺序执行，只有无依赖的任务可并行
3. 任务描述要具体可执行
4. 输入部分要包含该子任务需要的所有信息
5. 对于需要顺序执行的子任务，设置can_parallel: false
6. 通过execution_order字段指定建议的执行顺序
7. 确保每个子任务都有唯一的索引，用于依赖关系
}

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
			Description    string                 `json:"description"`
			Input          map[string]interface{} `json:"input"`
			Dependencies   []interface{}          `json:"dependencies"` // 改为interface{}兼容数字和字符串类型
			ExecutionOrder int                    `json:"execution_order"`
			CanParallel    bool                   `json:"can_parallel"`
		} `json:"subtasks"`
	}

	// 清理响应内容，只保留JSON部分
	jsonStr := extractJSON(response)
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("解析拆分结果失败: %w", err)
	}
	ts.log.Info("拆分结果: %+v", result)

	// 创建SubTask对象
	var subTasks []*domain.SubTask
	for i, st := range result.Subtasks {
		// 转换依赖为字符串数组，兼容数字类型的依赖
		var deps []string
		for _, dep := range st.Dependencies {
			deps = append(deps, fmt.Sprintf("%v", dep))
		}

		subTask := &domain.SubTask{
			Model: gorm.Model{
				ID: uint(i + 1), // 子任务ID从1开始，避免默认值0
			},
			Description:    st.Description,
			Status:         domain.StatusPending,
			Input:          st.Input,
			TaskID:         task.ID,
			Dependencies:   deps,
			ExecutionOrder: st.ExecutionOrder,
			CanParallel:    st.CanParallel,
			IsDAGNode:      len(st.Dependencies) > 0, // 如果有依赖关系，标记为DAG节点
		}

		// 如果LLM没有设置执行顺序，按默认顺序设置
		if subTask.ExecutionOrder == 0 && i > 0 {
			subTask.ExecutionOrder = i
		}

		subTasks = append(subTasks, subTask)
	}

	// 优化依赖关系：将文本索引转换为实际ID引用
	ts.optimizeDependencies(subTasks)

	ts.log.Info("任务拆分完成，共拆分为 %d 个子任务", len(subTasks))
	ts.log.Info("子任务依赖关系:")
	for i, st := range subTasks {
		if len(st.Dependencies) > 0 {
			ts.log.Info("  任务%d (顺序:%d, 并行:%v) 依赖: %v",
				i+1, st.ExecutionOrder, st.CanParallel, st.Dependencies)
		} else {
			ts.log.Info("  任务%d (顺序:%d, 并行:%v) 无依赖",
				i+1, st.ExecutionOrder, st.CanParallel)
		}
	}

	return subTasks, nil
}

// optimizeDependencies 优化依赖关系，将文本索引转换为实际ID引用
func (ts *TaskSplitter) optimizeDependencies(subTasks []*domain.SubTask) {
	// 创建索引映射
	indexMap := make(map[string]string)
	for i, task := range subTasks {
		// 使用任务索引作为键，支持多种格式
		indexMap[strconv.Itoa(i)] = fmt.Sprintf("task_%d", task.ID)
		indexMap[fmt.Sprintf("task_%d", i)] = fmt.Sprintf("task_%d", task.ID)
		indexMap[fmt.Sprintf("task_%d", i+1)] = fmt.Sprintf("task_%d", task.ID)
		indexMap[fmt.Sprintf("任务%d", i)] = fmt.Sprintf("task_%d", task.ID)
		indexMap[fmt.Sprintf("任务%d", i+1)] = fmt.Sprintf("task_%d", task.ID)
	}

	// 转换依赖关系
	for _, task := range subTasks {
		var normalizedDeps []string
		for _, dep := range task.Dependencies {
			// 清理依赖字符串
			dep = strings.TrimSpace(dep)
			dep = strings.Trim(dep, "\"")

			// 尝试映射
			if mapped, ok := indexMap[dep]; ok {
				normalizedDeps = append(normalizedDeps, mapped)
			} else {
				// 保持原样
				normalizedDeps = append(normalizedDeps, dep)
			}
		}
		task.Dependencies = normalizedDeps
	}
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