package repository

import (
	"orange-agent/domain"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	CreateUser(user *domain.User) error
	GetUserByTelegramId(telegramId int64) (*domain.User, error)
	UpdateUserModelName(telegramId int64, modelName string) error
}

// MemoryRepository 记忆仓储接口
type MemoryRepository interface {
	CreateMemory(memory *domain.Memory) error
	GetMemoryByUserId(userId uint) ([]domain.Memory, error)
	UpdateMemory(memory *domain.Memory) error
	GetMemoryByIdAndSize(memoryId uint, size int) ([]domain.Memory, error)
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
}
