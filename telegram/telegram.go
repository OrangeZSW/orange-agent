package telegram

import (
	"orange-agent/domain"
	"orange-agent/telegram/client"
	"orange-agent/telegram/interfaces"
)

func InitTelegram(config *domain.Telegram, answer interfaces.Ansewer) interfaces.Client {
	client := client.NewClient(answer)
	client.Init(config)
	return client
}
