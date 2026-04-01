package client

import (
	"context"
	"orange-agent/domain"
	"orange-agent/repository"
	"orange-agent/repository/resource"
	"orange-agent/telegram/interfaces"
	"orange-agent/telegram/manager"
	"orange-agent/utils"
	"orange-agent/utils/http"
	"orange-agent/utils/logger"

	"gopkg.in/telebot.v3"
)

type client struct {
	bot     *telebot.Bot
	log     *logger.Logger
	repo    *repository.Repositories
	manager interfaces.Manager
	answer  interfaces.Ansewer
}

func NewClient(answer interfaces.Ansewer) interfaces.Client {
	return &client{
		log:     logger.GetLogger(),
		repo:    resource.GetRepositories(),
		manager: manager.NewManager(),
		answer:  answer,
	}
}

func (c *client) Init(config *domain.Telegram) {

	pref := &telebot.Settings{
		Token:  config.BotToken,
		Client: http.GetHttpClient(config.Proxy),
	}
	bot, err := telebot.NewBot(*pref)
	if err != nil {
		c.log.Error("Failed to create bot: %v", err)
	}
	c.bot = bot
	c.listenMessage()
}

func (c *client) Start() {
	c.log.Info("Telegram Bot 已启动")
	c.bot.Start()
}

func (c *client) Stop() {
	c.log.Info("Telegram Bot 已停止")
	c.bot.Stop()
}

// 监听消息
func (c *client) listenMessage() {
	ctx := context.Background()
	c.bot.Handle(telebot.OnText, func(t telebot.Context) error {
		telegramId := t.Sender().ID
		name := t.Sender().Username
		user := c.manager.GetUser(telegramId, name)
		memory := &domain.Memory{
			UserId:       user.ID,
			UserQuestion: t.Text(),
		}
		ctx = utils.WithUser(ctx, user)
		c.repo.Memory.CreateMemory(memory)
		c.log.Info("Telegram收到消息: %s", t.Text())
		res := c.answer.TeleGramChat(ctx, user.ModelName, c.manager.GetMessage(user.ID, t.Text()))
		c.log.Info("Telegram发送消息: %s", res)
		memory.AgentAnswer = res
		c.repo.Memory.UpdateMemory(memory)
		err := t.Reply(res)
		if err != nil {
			c.log.Error("发送消息失败: %v", err)
		}
		return nil
	})
}

func (c *client) SendMessage(telegramId int64, text string) {
	c.log.Info("发送消息,userid:%d", telegramId)
	recipient := &telebot.User{
		ID: telegramId,
	}
	_, err := c.bot.Send(recipient, text)
	if err != nil {
		c.log.Error("发送消息失败: %v", err)
	}
}
