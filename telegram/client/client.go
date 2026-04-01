package client

import (
	"context"
	"orange-agent/domain"
	"orange-agent/repository"
	"orange-agent/repository/resource"
	"orange-agent/telegram/command"
	"orange-agent/telegram/interfaces"
	"orange-agent/telegram/manager"
	"orange-agent/utils"
	"orange-agent/utils/http"
	"orange-agent/utils/logger"
	"strings"

	"gopkg.in/telebot.v3"
)

type client struct {
	bot     *telebot.Bot
	log     *logger.Logger
	repo    *repository.Repositories
	manager interfaces.Manager
	answer  interfaces.Ansewer
	cmds    *command.CommandManager
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
	
	// 初始化命令管理器
	c.cmds = command.NewCommandManager(c.repo)
	
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
		messageText := t.Text()
		
		// 检查是否为快捷命令
		if strings.HasPrefix(messageText, "/") {
			c.log.Info("Telegram收到命令: %s", messageText)
			
			// 执行快捷命令
			result := c.cmds.Execute(ctx, t, user, messageText)
			
			// 记录到内存
			memory := &domain.Memory{
				UserId:       user.ID,
				UserQuestion: messageText,
				AgentAnswer:  result,
			}
			ctx = utils.WithUser(ctx, user)
			c.repo.Memory.CreateMemory(memory)
			c.repo.Memory.UpdateMemory(memory)
			
			// 发送命令结果
			c.log.Info("Telegram发送命令结果: %s", result)
			err := t.Reply(result, telebot.ModeMarkdown)
			if err != nil {
				c.log.Error("发送命令结果失败: %v", err)
			}
			return nil
		}
		
		// 原有AI助手消息处理逻辑
		memory := &domain.Memory{
			UserId:       user.ID,
			UserQuestion: messageText,
		}
		ctx = utils.WithUser(ctx, user)
		c.repo.Memory.CreateMemory(memory)
		c.log.Info("Telegram收到消息: %s", messageText)
		
		res := c.answer.TeleGramChat(ctx, user.ModelName, c.manager.GetMessage(user.ID, messageText))
		c.log.Info("Telegram发送消息: %s", res)
		
		memory.AgentAnswer = res
		c.repo.Memory.UpdateMemory(memory)
		
		err := t.Reply(res, telebot.ModeMarkdown)
		if err != nil {
			c.log.Error("发送消息失败: %v", err)
		}
		return nil
	})
	
	// 添加快捷命令帮助信息
	c.bot.Handle("/start", func(t telebot.Context) error {
		welcomeMsg := `🤖 *欢迎使用 Orange Agent!*

我是一个智能开发助手，可以帮助您：
• 📝 编写和重构代码
• 🔧 执行开发任务
• 📁 管理文件和项目
• 🗄️ 操作数据库
• 🤖 管理AI Agent

📋 *快捷命令*:
使用 /help 查看所有可用命令
使用 /status 查看系统状态

💡 *提示*:
直接发送消息与我对话，或使用快捷命令快速执行操作。`
		
		return t.Reply(welcomeMsg, telebot.ModeMarkdown)
	})
	
	c.bot.Handle("/help", func(t telebot.Context) error {
		result := c.cmds.Execute(context.Background(), t, nil, "/help")
		return t.Reply(result, telebot.ModeMarkdown)
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