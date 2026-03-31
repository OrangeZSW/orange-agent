// orchestrator/task_orchestrator.go
package orchestrator

import (
	"context"
	"fmt"
	"orange-agent/domain"
	"orange-agent/repository/factory"
	"orange-agent/task/analyzer"
	taskContext "orange-agent/task/context"
	"orange-agent/task/executor"
	"orange-agent/task/summarizer"
	"sync"
	"time"
)

type TaskOrchestrator struct {
	analyzer   *analyzer.TaskAnalyzer
	executor   *executor.TaskExecutor
	summarizer *summarizer.TaskSummarizer
	cm         *taskContext.ContextManager

	activeTasks map[uint]*domain.Task
	mu          sync.RWMutex
	repo        factory.Factory
}

func NewTaskOrchestrator(
	analyzer *analyzer.TaskAnalyzer,
	executor *executor.TaskExecutor,
	summarizer *summarizer.TaskSummarizer,
	cm *taskContext.ContextManager,
) *TaskOrchestrator {
	return &TaskOrchestrator{
		analyzer:    analyzer,
		executor:    executor,
		summarizer:  summarizer,
		cm:          cm,
		activeTasks: make(map[uint]*domain.Task),
		repo:        *factory.NewFactory(),
	}
}

// ProcessTask 处理总任务
func (to *TaskOrchestrator) ProcessTask(ctx context.Context, sessionID, taskDescription string) (*domain.Task, error) {
	// 创建总任务
	task := &domain.Task{
		SessionID:   sessionID,
		Description: taskDescription,
		Status:      domain.StatusPending,
		Subtasks:    []*domain.SubTask{},
	}

	to.mu.Lock()
	to.activeTasks[task.ID] = task
	to.mu.Unlock()

	defer func() {
		to.mu.Lock()
		delete(to.activeTasks, task.ID)
		to.mu.Unlock()
	}()

	// 步骤1: 分析并拆分任务
	task.Status = domain.StatusRunning
	subtasks, err := to.analyzer.AnalyzeAndSplit(taskDescription)
	if err != nil {
		task.Status = domain.StatusFailed
		return task, fmt.Errorf("failed to analyze task: %w", err)
	}

	task.Subtasks = subtasks

	// 步骤2: 执行子任务
	results := to.executor.ExecuteSubtasks(ctx, subtasks)

	// 步骤3: 总结任务
	summary, err := to.summarizer.Summarize(taskDescription, subtasks, results)
	if err != nil {
		task.Status = domain.StatusFailed
		task.Result = "Failed to generate summary: " + err.Error()
		return task, err
	}

	task.Result = summary
	task.Status = domain.StatusCompleted
	task.UpdatedAt = time.Now()

	return task, nil
}

// GetTaskStatus 获取任务状态
func (to *TaskOrchestrator) GetTaskStatus(taskID uint) (*domain.Task, error) {
	to.mu.RLock()
	defer to.mu.RUnlock()

	task, exists := to.activeTasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found")
	}

	return task, nil
}
