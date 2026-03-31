package interfaces

import (
	"orange-agent/domain"

	"github.com/tmc/langchaingo/llms"
)

type Client interface {
	Chat(modelName string, message []llms.MessageContent) string
}

type Agent interface {
	TeleGramChat(modelNmae string, message []llms.MessageContent, user *domain.User) string
}

type Manager interface {
	SaveCallRecord(message []llms.MessageContent, resp *llms.ContentResponse, agentConfig *domain.AgentConfig) error
	TeleGramSendMessage(text string)
}

type Telegram interface {
	SendTeleGramMessage(telegramId int64, text string)
}
