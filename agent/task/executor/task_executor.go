// executor/task_executor.go
package executor

import (
	"context"
	"fmt"
	taskContext "orange-agent/agent/task/context"
	"orange-agent/domain"
	"sync"
	"time"
)

type TaskExecutor struct {
	llmClient      LLMClient
	contextManager *taskContext.ContextManager
	workerCount    int
	resultChan     chan *domain.TaskResult
}

type LLMClient interface {
	ChatWithContext(ctx context.Context, taskCtx *domain.TaskContext, userMessage string) (string, int, error)
}

func NewTaskExecutor(llmClient LLMClient, cm *taskContext.ContextManager, workerCount int) *TaskExecutor {
	return &TaskExecutor{
		llmClient:      llmClient,
		contextManager: cm,
		workerCount:    workerCount,
		resultChan:     make(chan *domain.TaskResult, 100),
	}
}

// ExecuteSubtasks 并行执行子任务
func (te *TaskExecutor) ExecuteSubtasks(ctx context.Context, subtasks []*domain.SubTask) []*domain.TaskResult {
	var wg sync.WaitGroup
	results := make([]*domain.TaskResult, 0, len(subtasks))

	// 创建工作池
	workQueue := make(chan *domain.SubTask, len(subtasks))
	for i := 0; i < te.workerCount; i++ {
		wg.Add(1)
		go te.worker(ctx, &wg, workQueue)
	}

	// 发送任务
	for _, subtask := range subtasks {
		workQueue <- subtask
	}
	close(workQueue)

	// 等待完成
	go func() {
		wg.Wait()
		close(te.resultChan)
	}()

	// 收集结果
	for result := range te.resultChan {
		results = append(results, result)
	}

	return results
}

// worker 工作协程
func (te *TaskExecutor) worker(ctx context.Context, wg *sync.WaitGroup, workQueue <-chan *domain.SubTask) {
	defer wg.Done()

	for subtask := range workQueue {
		result := te.executeSubTask(ctx, subtask)
		te.resultChan <- result
	}
}

// executeSubTask 执行单个子任务
func (te *TaskExecutor) executeSubTask(ctx context.Context, subtask *domain.SubTask) *domain.TaskResult {
	startTime := time.Now()
	result := &domain.TaskResult{
		SubTaskID: subtask.ID,
		Success:   false,
	}

	// 更新状态
	subtask.Status = domain.StatusRunning

	// 为子任务创建独立上下文
	systemPrompt := "You are an AI assistant. Complete the following subtask accurately and concisely."
	if subtask.Context != nil && subtask.Context.SystemPrompt != "" {
		systemPrompt = subtask.Context.SystemPrompt
	}

	taskContext := te.contextManager.CreateContext(subtask.ID, systemPrompt)

	// 添加系统提示
	te.contextManager.AddMessage(subtask.ID, "system", systemPrompt)

	// 构建执行提示
	executionPrompt := fmt.Sprintf("Execute the following subtask:\n\n%s\n\nProvide a clear and concise response.", subtask.Description)

	// 调用LLM
	output, tokenUsed, err := te.llmClient.ChatWithContext(ctx, taskContext, executionPrompt)

	result.ExecutionMs = time.Since(startTime).Milliseconds()
	result.TokenUsed = tokenUsed

	if err != nil {
		result.Error = err.Error()
		subtask.Status = domain.StatusFailed
		subtask.Error = err.Error()
		return result
	}

	// 更新子任务
	subtask.Output = output
	subtask.Status = domain.StatusCompleted
	result.Success = true
	result.Output = output

	// 任务完成后，可以选择清理上下文
	// te.contextManager.ClearContext(subtask.ID)

	return result
}
