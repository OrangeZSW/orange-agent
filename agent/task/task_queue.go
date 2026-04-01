package task

import (
	"sync"
	"orange-agent/domain"
)

// TaskQueue 任务队列，管理待执行的子任务
type TaskQueue struct {
	tasks chan *domain.SubTask
	mu    sync.Mutex
	closed bool
}

// NewTaskQueue 创建新的任务队列
func NewTaskQueue(bufferSize int) *TaskQueue {
	return &TaskQueue{
		tasks: make(chan *domain.SubTask, bufferSize),
	}
}

// Enqueue 将子任务加入队列
func (tq *TaskQueue) Enqueue(subTask *domain.SubTask) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	
	if tq.closed {
		return ErrQueueClosed
	}
	
	subTask.Status = domain.StatusPending
	tq.tasks <- subTask
	return nil
}

// Dequeue 从队列中取出子任务
func (tq *TaskQueue) Dequeue() (*domain.SubTask, bool) {
	subTask, ok := <-tq.tasks
	if ok {
		subTask.Status = domain.StatusRunning
	}
	return subTask, ok
}

// Size 获取队列当前大小
func (tq *TaskQueue) Size() int {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	return len(tq.tasks)
}

// Close 关闭队列
func (tq *TaskQueue) Close() {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	
	if !tq.closed {
		tq.closed = true
		close(tq.tasks)
	}
}

// IsClosed 检查队列是否已关闭
func (tq *TaskQueue) IsClosed() bool {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	return tq.closed
}
