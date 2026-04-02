package telegram

import (
	"orange-agent/domain"
	"orange-agent/telegram/client"
	"orange-agent/telegram/interfaces"
	"sync"
)

var (
	instance interfaces.Telegram
	once     sync.Once
)

type telegram struct {
	client interfaces.Client
}

// NewTelegram 创建Telegram实例（单例模式）
func NewTelegram() interfaces.Telegram {
	once.Do(func() {
		instance = &telegram{}
	})
	return instance
}

// Init 初始化Telegram机器人
func (t *telegram) Init(config *domain.Telegram, handler interfaces.MessageHandler) interfaces.Client {
	c := client.NewClient(handler)
	c.Init(config)
	t.client = c
	return c
}

// Start 启动机器人
func (t *telegram) Start() {
	if t.client != nil {
		t.client.Start()
	}
}

// Stop 停止机器人
func (t *telegram) Stop() {
	if t.client != nil {
		t.client.Stop()
	}
}

// SendMessage 发送消息给指定用户
func (t *telegram) SendMessage(telegramId int64, text string) {
	if t.client != nil {
		t.client.SendMessage(telegramId, text)
	}
}
