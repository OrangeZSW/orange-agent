package interfaces

import (
	"context"
	"orange-agent/domain"

	"github.com/tmc/langchaingo/llms"
)

// Client 定义Agent客户端接口
type Client interface {
	// Chat 与AI模型对话
	Chat(modelName string, message []llms.MessageContent) string
}

// Agent 定义Agent核心接口
type Agent interface {
	Handle(ctx context.Context, modelName string, messages []llms.MessageContent) string
	// Chat 处理通用对话
	Chat(ctx context.Context, messages []domain.Message) string
}

// Manager 定义Agent管理器接口
type Manager interface {
	// SaveCallRecord 保存调用记录
	SaveCallRecord(message []llms.MessageContent, resp *llms.ContentResponse, agentConfig *domain.AgentConfig) error
	// SendMessage 通过Telegram发送消息
	SendMessage(text string)
	// SystemPrompt 获取系统提示词
	SystemPrompt() []llms.MessageContent
}

// Telegram 定义Telegram接口
type Telegram interface {
	// SendMessage 发送消息给指定用户
	SendMessage(telegramId int64, text string)
}

// AgentManager 定义Agent管理器接口
type AgentManager interface {
	// GetDefaultAgent 获取默认Agent
	GetDefaultAgent() (Agent, error)
	// GetAgentByName 根据名称获取Agent
	GetAgentByName(name string) (Agent, error)
}
