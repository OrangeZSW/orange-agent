package interfaces

import (
	"orange-agent/domain"

	"github.com/tmc/langchaingo/llms"
)

type Telegram interface {
	SendTeleGramMessage(telegramId int64, text string)
	InitTelegram(config *domain.Telegram, answer Ansewer) Client
}

type Manager interface {
	GetUser(telegramId int64, name string) *domain.User
	GetMessage(id uint, question string) []llms.MessageContent
}

type Client interface {
	Init(config *domain.Telegram)
	Start()
	SendMessage(telegramId int64, text string)
}

type Ansewer interface {
	TeleGramChat(modelNmae string, message []llms.MessageContent, user *domain.User) string
}
