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
	"orange-agent/utils/logger"
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
	log         *logger.Logger
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
		log:         logger.GetLogger(),
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
	err := to.repo.TaskRepo.CreateTask(task)

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
	to.log.Info("分析任务成功，创建子任务：%v", subtasks)

	for _, subtask := range subtasks {
		subtask.TaskID = task.ID
		to.repo.SubTaskRepo.CreateSubTask(subtask)
	}
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

	to.repo.TaskRepo.UpdateTask(task)

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
