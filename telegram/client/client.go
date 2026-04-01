package client

import (
	"context"
	"fmt"
	"orange-agent/domain"
	"orange-agent/repository"
	"orange-agent/repository/resource"
	"orange-agent/telegram/command"
	"orange-agent/telegram/interfaces"
	"orange-agent/telegram/manager"
	"orange-agent/telegram/ui"
	"orange-agent/utils"
	"orange-agent/utils/http"
	"orange-agent/utils/logger"

	"gopkg.in/telebot.v3"
)

const maxMessageLength = 4096 // Telegram单条消息最大长度限制

type client struct {
	bot     *telebot.Bot
	log     *logger.Logger
	repo    *repository.Repositories
	manager interfaces.Manager
	answer  interfaces.Ansewer
	cmds    *command.CommandManager
	ui      *ui.UIManager // 添加UI管理器
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

	// 初始化UI管理器
	c.ui = ui.NewUIManager(c.cmds, c.answer)

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
	// 处理所有文本消息
	c.bot.Handle(telebot.OnText, func(t telebot.Context) error {
		telegramId := t.Sender().ID
		name := t.Sender().Username
		user := c.manager.GetUser(telegramId, name)
		messageText := t.Text()

		c.log.Info("Telegram收到消息: %s, 用户: %d", messageText, telegramId)

		// 先将用户信息存入上下文，再传给后续处理
		ctx := utils.WithUser(context.Background(), user)

		// 使用UI管理器处理消息
		result, menu, err := c.ui.HandleMessage(ctx, t, user, messageText)
		if err != nil {
			c.log.Error("UI处理消息失败: %v", err)
			return t.Reply("❌ 处理消息时出错", telebot.ModeMarkdown)
		}

		// 记录到内存
		memory := &domain.Memory{
			UserId:       user.ID,
			UserQuestion: messageText,
			AgentAnswer:  result,
		}
		c.repo.Memory.CreateMemory(memory)

		// 发送响应
		if menu != nil {
			c.log.Info("发送带菜单的响应: %s", result)
			return c.ui.SendMessageWithMenu(t, result, menu)
		} else {
			c.log.Info("发送响应: %s", result)
			c.repo.Memory.UpdateMemory(memory)
			return c.sendLongMessage(t, result)
		}
	})

	// 处理按钮回调
	c.bot.Handle(telebot.OnCallback, func(t telebot.Context) error {
		telegramId := t.Sender().ID
		name := t.Sender().Username
		user := c.manager.GetUser(telegramId, name)
		data := t.Data()

		c.log.Info("Telegram收到按钮回调: %s, 用户: %d", data, telegramId)

		// 先将用户信息存入上下文，再传给后续处理
		ctx := utils.WithUser(context.Background(), user)

		// 使用UI管理器处理回调
		result, menu, err := c.ui.HandleCallback(ctx, t, user, data)
		if err != nil {
			c.log.Error("UI处理回调失败: %v", err)
			return t.Respond(&telebot.CallbackResponse{
				Text: "❌ 处理回调时出错",
			})
		}

		// 记录回调
		memory := &domain.Memory{
			UserId:       user.ID,
			UserQuestion: fmt.Sprintf("[按钮] %s", data),
			AgentAnswer:  result,
		}
		c.repo.Memory.CreateMemory(memory)
		c.repo.Memory.UpdateMemory(memory)

		// 发送响应
		if menu != nil {
			c.log.Info("发送带菜单的按钮响应: %s", result)
			return c.ui.SendMessageWithMenu(t, result, menu)
		} else {
			c.log.Info("发送按钮响应: %s", result)
			return c.sendLongMessage(t, result)
		}
	})

	// 处理 /start 命令 - 显示主菜单
	c.bot.Handle("/start", func(t telebot.Context) error {
		welcomeMsg := `🤖 *欢迎使用 Orange Agent!*

我是一个智能开发助手，可以通过点击按钮快速执行操作：

📋 *点击按钮即可执行功能：*
• 📁 文件管理 - 查看、读取、搜索文件
• 🔧 Git操作 - 提交、推送代码
• 🏗️ 项目管理 - 构建、测试项目
• 🗄️ 数据库 - 查询和执行SQL
• 🤖 Agent管理 - 配置AI Agent
• ⚙️ 系统工具 - 监控、搜索、API测试

💡 *提示*:
- 直接点击按钮即可执行操作
- 某些操作需要输入参数，按照提示输入即可
- 所有操作历史会被记录`

		menu := c.ui.GetMenuManager() // 使用公共方法获取菜单
		return t.Reply(welcomeMsg, menu, telebot.ModeMarkdown)
	})

	// 处理 /help 命令
	c.bot.Handle("/help", func(t telebot.Context) error {
		telegramId := t.Sender().ID
		name := t.Sender().Username
		user := c.manager.GetUser(telegramId, name)
		ctx := utils.WithUser(context.Background(), user)

		result := c.cmds.Execute(ctx, t, user, "/help")

		// 返回结果和主菜单
		menu := c.ui.GetMenuManager() // 使用公共方法获取菜单
		return t.Reply(result, menu, telebot.ModeMarkdown)
	})

	// 添加快捷命令的别名处理
	c.bot.Handle("/list", func(t telebot.Context) error {
		ctx := context.Background()
		result := c.cmds.Execute(ctx, t, nil, "/list")
		menu := c.ui.GetFileMenu() // 使用公共方法获取文件菜单
		return t.Reply(result, menu, telebot.ModeMarkdown)
	})

	c.bot.Handle("/status", func(t telebot.Context) error {
		ctx := context.Background()
		result := c.cmds.Execute(ctx, t, nil, "/status")
		menu := c.ui.GetMenuManager() // 使用公共方法获取菜单
		return t.Reply(result, menu, telebot.ModeMarkdown)
	})
}

// 发送长消息，自动拆分超过长度限制的内容
func (c *client) sendLongMessage(t telebot.Context, text string) error {
	if len(text) <= maxMessageLength {
		return t.Reply(text, telebot.ModeMarkdown)
	}

	// 拆分文本为多个块
	var chunks []string
	for i := 0; i < len(text); i += maxMessageLength {
		end := i + maxMessageLength
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[i:end])
	}

	// 逐个发送消息块
	for i, chunk := range chunks {
		var err error
		if i == len(chunks)-1 {
			err = t.Reply(chunk, telebot.ModeMarkdown)
		} else {
			err = t.Reply(chunk+"\n...(continued)", telebot.ModeMarkdown)
		}
		if err != nil {
			c.log.Error("发送消息块失败: %v", err)
			return err
		}
	}
	return nil
}

func (c *client) SendMessage(telegramId int64, text string) {
	c.log.Info("发送消息,userid:%d", telegramId)
	recipient := &telebot.User{
		ID: telegramId,
	}

	if len(text) <= maxMessageLength {
		_, err := c.bot.Send(recipient, text, telebot.ModeMarkdown)
		if err != nil {
			c.log.Error("发送消息失败: %v", err)
		}
		return
	}

	// 拆分长消息发送
	var chunks []string
	for i := 0; i < len(text); i += maxMessageLength {
		end := i + maxMessageLength
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[i:end])
	}

	// 逐个发送消息块
	for i, chunk := range chunks {
		sendText := chunk
		if i != len(chunks)-1 {
			sendText += "\n...(下一部分)"
		}
		_, err := c.bot.Send(recipient, sendText, telebot.ModeMarkdown)
		if err != nil {
			c.log.Error("发送消息块失败: %v", err)
			return
		}
	}
}
