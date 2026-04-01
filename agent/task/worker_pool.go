package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	"orange-agent/domain"
	"orange-agent/utils/logger"
)

// WorkerPool 工作池，并发执行子任务
type WorkerPool struct {
	workerCount    int
	taskQueue      *TaskQueue
	contextManager *ContextManager
	resultChan     chan *domain.SubTask
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	taskChat       TaskChat
}

// NewWorkerPool 创建新的工作池
func NewWorkerPool(
	workerCount int,
	taskQueue *TaskQueue,
	contextManager *ContextManager,
	taskChat TaskChat,
) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		workerCount:    workerCount,
		taskQueue:      taskQueue,
		contextManager: contextManager,
		resultChan:     make(chan *domain.SubTask, 100),
		ctx:            ctx,
		cancel:         cancel,
		taskChat:       taskChat,
	}
}

// Start 启动工作池
func (wp *WorkerPool) Start() {
	logger.Info("启动工作池，Worker数量: %d", wp.workerCount)

	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// Stop 停止工作池
func (wp *WorkerPool) Stop() {
	logger.Info("停止工作池")
	wp.cancel()
	wp.wg.Wait()
	close(wp.resultChan)
}

// GetResultChan 获取结果通道
func (wp *WorkerPool) GetResultChan() <-chan *domain.SubTask {
	return wp.resultChan
}

// worker 工作协程
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	logger.Debug("Worker %d 启动", id)

	for {
		select {
		case <-wp.ctx.Done():
			logger.Debug("Worker %d 收到停止信号", id)
			return
		case subTask, ok := <-wp.taskQueue.tasks:
			if !ok {
				logger.Debug("Worker %d 任务队列已关闭", id)
				return
			}

			logger.Info("Worker %d 开始执行子任务: %s", id, subTask.Description)
			wp.executeSubTask(subTask)

			// 发送结果
			wp.resultChan <- subTask
		}
	}
}

// executeSubTask 执行单个子任务
func (wp *WorkerPool) executeSubTask(subTask *domain.SubTask) {
	startTime := time.Now()
	defer func() {
		executionMs := time.Since(startTime).Milliseconds()
		if r := recover(); r != nil {
			errMsg := fmt.Sprintf("子任务执行恐慌: %v", r)
			logger.Error(errMsg)
			subTask.Status = domain.StatusFailed
			subTask.Error = errMsg
		}
		logger.Info("子任务执行完成，耗时: %dms", executionMs)
	}()

	// 创建独立的任务上下文
	taskCtx := wp.contextManager.CreateTaskContext(
		subTask.ID,
		"You are a helpful assistant that executes tasks efficiently.",
	)
	subTask.Context = taskCtx
	// 构建任务提示
	prompt := wp.buildTaskPrompt(subTask)

	// 添加用户消息到上下文
	wp.contextManager.AddMessage(subTask.ID, "user", prompt, len(prompt))

	// 执行任务
	messages := []domain.Message{
		{Role: "system", Content: taskCtx.SystemPrompt},
		{Role: "user", Content: prompt},
	}

	response := wp.taskChat.TaskChat(wp.ctx, messages)

	// 添加助手响应到上下文
	wp.contextManager.AddMessage(subTask.ID, "assistant", response, len(response))

	// 更新子任务状态
	subTask.Status = domain.StatusCompleted
	subTask.Output = response

	logger.Info("子任务执行成功: %s", subTask.Description)
}

// buildTaskPrompt 构建任务提示
func (wp *WorkerPool) buildTaskPrompt(subTask *domain.SubTask) string {
	prompt := fmt.Sprintf("请执行以下任务：\n\n任务描述：%s\n\n", subTask.Description)

	if len(subTask.Input) > 0 {
		prompt += "输入信息：\n"
		for key, value := range subTask.Input {
			prompt += fmt.Sprintf("- %s: %v\n", key, value)
		}
	}

	prompt += "\n请详细完成该任务，并返回结果。"
	return prompt
}
