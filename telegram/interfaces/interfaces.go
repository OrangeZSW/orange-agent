package interfaces

import (
	"orange-agent/domain"

	"github.com/tmc/langchaingo/llms"
)

type Telegram interface {
	Init(config *domain.Telegram)
	Start()
	Stop()
	SendMessage(telegramId int64, text string)
}

type Manager interface {
	GetUser(telegramId int64, name string) *domain.User
	GetMessage(id uint, question string) []llms.MessageContent
}

type Client interface {
	Init(config *domain.Telegram)
	Start()
}

type Ansewer interface {
	TeleGramChat(modelNmae string, message []llms.MessageContent, user *domain.User) string
}
