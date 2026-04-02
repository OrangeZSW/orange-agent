package task

import (
	"fmt"
	"orange-agent/domain"
	"sync"
)

// ContextManager 管理所有任务的上下文
type ContextManager struct {
	sessionStore    *SessionStore
	contextRegistry *TaskContextRegistry
}

// NewContextManager 创建新的上下文管理器
func NewContextManager() *ContextManager {
	return &ContextManager{
		sessionStore:    NewSessionStore(),
		contextRegistry: NewTaskContextRegistry(),
	}
}

// CreateTaskContext 创建新的任务上下文
func (cm *ContextManager) CreateTaskContext(subTaskID uint, systemPrompt string) *domain.TaskContext {
	ctx := &domain.TaskContext{
		SystemPrompt: systemPrompt,
		Messages:     make([]domain.Message, 0),
		TokenCount:   0,
		Metadata:     make(map[string]interface{}),
	}
	cm.contextRegistry.Register(subTaskID, ctx)
	return ctx
}

// GetTaskContext 获取指定子任务的上下文
func (cm *ContextManager) GetTaskContext(subTaskID uint) (*domain.TaskContext, bool) {
	return cm.contextRegistry.Get(subTaskID)
}

// GetContext 实现ContextManagerInterface接口
func (cm *ContextManager) GetContext(taskID uint) *domain.TaskContext {
	ctx, _ := cm.GetTaskContext(taskID)
	return ctx
}

// AddMessage 添加消息到指定上下文
func (cm *ContextManager) AddMessage(subTaskID uint, role, content string, tokenCount int) {
	if ctx, exists := cm.contextRegistry.Get(subTaskID); exists {
		ctx.Messages = append(ctx.Messages, domain.Message{
			Role:    role,
			Content: content,
		})
		ctx.TokenCount += tokenCount
	}
}

// CompressContext 压缩上下文（当token超过限制时）
func (cm *ContextManager) CompressContext(subTaskID uint, maxTokens int) error {
	ctx, exists := cm.contextRegistry.Get(subTaskID)
	if !exists {
		return fmt.Errorf("context for subTask %d not found", subTaskID)
	}
	if ctx.TokenCount <= maxTokens {
		return nil
	}
	// 压缩逻辑：保留系统prompt和最近的消息直到token数低于限制
	newMessages := []domain.Message{
		{Role: "system", Content: ctx.SystemPrompt},
	}
	newTokenCount := len(ctx.SystemPrompt)
	// 从后往前添加历史消息
	for i := len(ctx.Messages) - 1; i >= 0; i-- {
		msg := ctx.Messages[i]
		msgTokens := len(msg.Content)
		if newTokenCount + msgTokens > maxTokens {
			break
		}
		newMessages = append(newMessages, msg)
		newTokenCount += msgTokens
	}
	// 反转消息恢复正确顺序
	for i, j := 1, len(newMessages)-1; i < j; i, j = i+1, j-1 {
		newMessages[i], newMessages[j] = newMessages[j], newMessages[i]
	}
	ctx.Messages = newMessages
	ctx.TokenCount = newTokenCount
	return nil
}

// SessionStore 存储会话信息
type SessionStore struct {
	sessions map[string]*domain.Task
	mu       sync.RWMutex
}

// NewSessionStore 创建新的会话存储
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*domain.Task),
	}
}

// SaveSession 保存会话
func (ss *SessionStore) SaveSession(sessionID string, task *domain.Task) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.sessions[sessionID] = task
}

// GetSession 获取会话
func (ss *SessionStore) GetSession(sessionID string) (*domain.Task, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	task, exists := ss.sessions[sessionID]
	return task, exists
}

// TaskContextRegistry 注册和管理所有任务上下文
type TaskContextRegistry struct {
	contexts map[uint]*domain.TaskContext
	mu       sync.RWMutex
}

// NewTaskContextRegistry 创建新的任务上下文注册表
func NewTaskContextRegistry() *TaskContextRegistry {
	return &TaskContextRegistry{
		contexts: make(map[uint]*domain.TaskContext),
	}
}

// Register 注册任务上下文
func (tcr *TaskContextRegistry) Register(subTaskID uint, ctx *domain.TaskContext) {
	tcr.mu.Lock()
	defer tcr.mu.Unlock()
	tcr.contexts[subTaskID] = ctx
}

// Get 获取任务上下文
func (tcr *TaskContextRegistry) Get(subTaskID uint) (*domain.TaskContext, bool) {
	tcr.mu.RLock()
	defer tcr.mu.RUnlock()
	ctx, exists := tcr.contexts[subTaskID]
	return ctx, exists
}

// Remove 移除任务上下文
func (tcr *TaskContextRegistry) Remove(subTaskID uint) {
	tcr.mu.Lock()
	defer tcr.mu.Unlock()
	delete(tcr.contexts, subTaskID)
}
