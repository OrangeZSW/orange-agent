// bot/bot.go
package telegram

import (
	"net/http"
	"net/url"
	"orange-agent/config"
	"orange-agent/utils/logger"
	"time"

	tele "gopkg.in/telebot.v3"
)

type TelegramBot struct {
	Bot            *tele.Bot
	Config         *config.Telegram
	handlerCommand *HandlerCommand
	log            *logger.Logger
	HandlerText    *HandlerText
}

func NewTelegramBot(config *config.Telegram) *TelegramBot {
	b := NewTelegramBotWithProxy(config)
	telegram := &TelegramBot{
		Config: config,
		log:    logger.GetLogger(),
		Bot:    b,
	}
	telegram.handlerCommand = NewHandlerCommand(telegram)
	telegram.HandlerText = NewHandlerText(telegram)
	return telegram
}

func (tb *TelegramBot) registerHandlers() {
	tb.Bot.Handle(tele.OnText, tb.handleMessage)
}

func (tb *TelegramBot) handleMessage(c tele.Context) error {
	userID := c.Sender().ID
	message := c.Text()

	tb.log.Info("收到用户 %d 消息: %s", userID, message)

	// 调用你的 AI 处理逻辑
	// response := oz_agent.Answer(...)

	response := "正在处理你的消息..." // 替换为实际 AI 回复

	return c.Send(response)
}

// Start 启动 Bot
func (tb *TelegramBot) Start() {
	logger.Info("Telegram Bot 已启动")
	tb.Bot.Start()
}

// Stop 停止 Bot
func (tb *TelegramBot) Stop() {
	tb.Bot.Stop()
	logger.Info("Telegram Bot 已停止")
}

func NewTelegramBotWithProxy(config *config.Telegram) *tele.Bot {
	// 配置代理
	proxy, err := url.Parse(config.Proxy)
	if err != nil {
		panic(err)
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxy),
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	pref := tele.Settings{
		Token:  config.BotToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
		Client: client,
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		panic(err)
	}
	return b
}
