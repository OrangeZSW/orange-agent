package repository

import (
	"orange-agent/domain"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	CreateUser(user *domain.User) error
	GetUserByTelegramId(telegramId int64) (*domain.User, error)
	UpdateUserModelName(telegramId int64, modelName string) error
	GetUserById(id uint) (*domain.User, error)
}

// MemoryRepository 记忆仓储接口
type MemoryRepository interface {
	CreateMemory(memory *domain.Memory) error
	GetMemoryByUserId(userId uint) ([]domain.Memory, error)
	UpdateMemory(memory *domain.Memory) error
	GetMemoryByUserIdAndLimit(userId uint, limit int) ([]domain.Memory, error)
}

// AgentConfigRepository Agent配置仓储接口
type AgentConfigRepository interface {
	GetAgentConfigByModel(modelName string) (*domain.AgentConfig, error)
	GetAgentConfigByName(name string) (*domain.AgentConfig, error)
	CreateAgentConfig(agentConfig *domain.AgentConfig) error
	UpdateAgentConfig(agentConfig *domain.AgentConfig) error
	GetAgentConfigById(id uint) (*domain.AgentConfig, error)
	GetAllAgentConfig() ([]domain.AgentConfig, error)
	DeleteAgentConfig(agentConfig *domain.AgentConfig) error
}

// AgentCallRecordRepository Agent调用记录仓储接口
type AgentCallRecordRepository interface {
	CreateAgentCallRecord(agentCallRecord *domain.CallRecord) error
	GetAgentCallRecordByAgentName(agentName string) ([]domain.CallRecord, error)
	SelectByMemoryId(memoryId uint) ([]domain.CallRecord, error)
}

// Task
type TaskRepository interface {
	CreateTask(task *domain.Task) error
	UpdateTask(task *domain.Task) error
	GetTaskById(id uint) (*domain.Task, error)
}

// SubTask
type SubTaskRepository interface {
	CreateSubTask(subTask *domain.SubTask) error
	GetSubTaskByTaskId(taskId uint) ([]domain.SubTask, error)
	UpdateSubTask(subTask *domain.SubTask) error
	GetSubTaskById(id uint) (*domain.SubTask, error)
}

// TaskResult
type TaskResultRepository interface {
	CreateTaskResult(taskResult *domain.TaskResult) error
	GetTaskResultBySubTaskId(subTaskId uint) (*domain.TaskResult, error)
	UpdateTaskResult(taskResult *domain.TaskResult) error
	GetTaskResultById(id uint) (*domain.TaskResult, error)
}
