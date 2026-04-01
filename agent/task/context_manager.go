package task

import (
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
func (cm *ContextManager) CompressContext(subTaskID uint, summary string) {
	if ctx, exists := cm.contextRegistry.Get(subTaskID); exists {
		// 保留系统prompt和摘要
		ctx.Messages = []domain.Message{
			{Role: "system", Content: ctx.SystemPrompt},
			{Role: "assistant", Content: "Previous conversation summary: " + summary},
		}
		// 重置token计数（这里简化处理，实际应该计算摘要的token数）
		ctx.TokenCount = 0
	}
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
