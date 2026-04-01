package interfaces

import (
	"context"
	"orange-agent/domain"

	"github.com/tmc/langchaingo/llms"
)

type Client interface {
	Chat(modelName string, message []llms.MessageContent) string
}

type Agent interface {
	TeleGramChat(modelNmae string, message []llms.MessageContent, user *domain.User) string
	Chat(ctx context.Context, messages []domain.Message) (*domain.Message, error)
}

type Manager interface {
	SaveCallRecord(message []llms.MessageContent, resp *llms.ContentResponse, agentConfig *domain.AgentConfig) error
	TeleGramSendMessage(text string)
}

type Telegram interface {
	SendTeleGramMessage(telegramId int64, text string)
}

type AgentManager interface {
	GetDefaultAgent() (Agent, error)
	GetAgentByName(name string) (Agent, error)
}
