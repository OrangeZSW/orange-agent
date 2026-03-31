package telegram

import (
	"orange-agent/domain"
	"orange-agent/telegram/client"
	"orange-agent/telegram/interfaces"
	"sync"
)

var (
	teleGram interfaces.Telegram
	once     sync.Once
)

type telegram struct {
	client interfaces.Client
}

func NewTelegram() interfaces.Telegram {
	once.Do(func() {
		teleGram = &telegram{}
	})
	return teleGram
}

func (t *telegram) InitTelegram(config *domain.Telegram, answer interfaces.Ansewer) interfaces.Client {
	client := client.NewClient(answer)
	client.Init(config)
	return client
}

func (t *telegram) SendTeleGramMessage(telegramId int64, text string) {
	t.client.SendMessage(telegramId, text)
}

func GetTelegram() interfaces.Telegram {
	return NewTelegram()
}
