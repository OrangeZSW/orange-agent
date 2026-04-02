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

// MaxMessageLength Telegram单条消息最大长度限制
const MaxMessageLength = 4096

type client struct {
	bot     *telebot.Bot
	log     *logger.Logger
	repo    *repository.Repositories
	mgr     interfaces.Manager
	handler interfaces.MessageHandler
	cmds    *command.CommandManager
	ui      *ui.UIManager
}

// NewClient 创建Telegram客户端
func NewClient(handler interfaces.MessageHandler) interfaces.Client {
	return &client{
		log:     logger.GetLogger(),
		repo:    resource.GetRepositories(),
		mgr:     manager.NewManager(),
		handler: handler,
	}
}

// Init 初始化客户端
func (c *client) Init(config *domain.Telegram) {
	pref := &telebot.Settings{
		Token:  config.BotToken,
		Client: http.GetHttpClient(config.Proxy),
	}
	bot, err := telebot.NewBot(*pref)
	if err != nil {
		c.log.Error("创建Bot失败: %v", err)
		return
	}
	c.bot = bot

	// 初始化命令管理器
	c.cmds = command.NewCommandManager(c.repo)

	// 初始化UI管理器
	c.ui = ui.NewUIManager(c.cmds, c.handler)

	c.registerHandlers()
}

// Start 启动机器人
func (c *client) Start() {
	if c.bot == nil {
		c.log.Error("Bot未初始化")
		return
	}
	c.log.Info("Telegram Bot 已启动")
	c.bot.Start()
}

// Stop 停止机器人
func (c *client) Stop() {
	if c.bot != nil {
		c.log.Info("Telegram Bot 已停止")
		c.bot.Stop()
	}
}

// SendMessage 发送消息给指定用户
func (c *client) SendMessage(telegramId int64, text string) {
	if c.bot == nil {
		return
	}

	recipient := &telebot.User{ID: telegramId}
	chunks := splitMessage(text, MaxMessageLength)

	for i, chunk := range chunks {
		sendText := chunk
		if i < len(chunks)-1 {
			sendText += "\n...(下一部分)"
		}
		if _, err := c.bot.Send(recipient, sendText, telebot.ModeMarkdown); err != nil {
			c.log.Error("发送消息失败: %v", err)
		}
	}
}

// registerHandlers 注册消息处理器
func (c *client) registerHandlers() {
	c.bot.Handle(telebot.OnText, c.handleTextMessage)
	c.bot.Handle(telebot.OnCallback, c.handleCallback)
	c.bot.Handle("/start", c.handleStart)
	c.bot.Handle("/help", c.handleHelp)
	c.bot.Handle("/list", c.handleList)
	c.bot.Handle("/status", c.handleStatus)
}

// handleTextMessage 处理文本消息
func (c *client) handleTextMessage(t telebot.Context) error {
	user := c.getUser(t)
	messageText := t.Text()
	c.log.Info("Telegram收到消息: %s, 用户: %d", messageText, user.TelegramId)

	ctx := utils.WithUser(context.Background(), user)
	result, menu, err := c.ui.HandleMessage(ctx, t, user, messageText)
	if err != nil {
		c.log.Error("处理消息失败: %v", err)
		return t.Reply("❌ 处理消息时出错", telebot.ModeMarkdown)
	}

	c.saveMemory(user, messageText, result)

	if menu != nil {
		return c.ui.SendMessageWithMenu(t, result, menu)
	}
	return c.replyLongMessage(t, result)
}

// handleCallback 处理按钮回调
func (c *client) handleCallback(t telebot.Context) error {
	user := c.getUser(t)
	data := t.Data()
	c.log.Info("Telegram收到按钮回调: %s, 用户: %d", data, user.TelegramId)

	ctx := utils.WithUser(context.Background(), user)
	result, menu, err := c.ui.HandleCallback(ctx, t, user, data)
	if err != nil {
		c.log.Error("处理回调失败: %v", err)
		return t.Respond(&telebot.CallbackResponse{Text: "❌ 处理回调时出错"})
	}

	c.saveMemory(user, fmt.Sprintf("[按钮] %s", data), result)

	if menu != nil {
		return c.ui.SendMessageWithMenu(t, result, menu)
	}
	return c.replyLongMessage(t, result)
}

// handleStart 处理/start命令
func (c *client) handleStart(t telebot.Context) error {
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

	return t.Reply(welcomeMsg, c.ui.GetMainMenu(), telebot.ModeMarkdown)
}

// handleHelp 处理/help命令
func (c *client) handleHelp(t telebot.Context) error {
	user := c.getUser(t)
	ctx := utils.WithUser(context.Background(), user)
	result := c.cmds.Execute(ctx, t, user, "/help")
	return t.Reply(result, c.ui.GetMainMenu(), telebot.ModeMarkdown)
}

// handleList 处理/list命令
func (c *client) handleList(t telebot.Context) error {
	user := c.getUser(t)
	ctx := utils.WithUser(context.Background(), user)
	result := c.cmds.Execute(ctx, t, user, "/list")
	return t.Reply(result, c.ui.GetFileMenu(), telebot.ModeMarkdown)
}

// handleStatus 处理/status命令
func (c *client) handleStatus(t telebot.Context) error {
	user := c.getUser(t)
	ctx := utils.WithUser(context.Background(), user)
	result := c.cmds.Execute(ctx, t, user, "/status")
	return t.Reply(result, c.ui.GetMainMenu(), telebot.ModeMarkdown)
}

// getUser 获取用户信息
func (c *client) getUser(t telebot.Context) *domain.User {
	return c.mgr.GetUser(t.Sender().ID, t.Sender().Username)
}

// saveMemory 保存对话记录
func (c *client) saveMemory(user *domain.User, question, answer string) {
	memory := &domain.Memory{
		UserId:       user.ID,
		UserQuestion: question,
		AgentAnswer:  answer,
	}
	if err := c.repo.Memory.CreateMemory(memory); err != nil {
		c.log.Warn("保存记忆失败: %v", err)
	}
}

// replyLongMessage 回复长消息（自动拆分）
func (c *client) replyLongMessage(t telebot.Context, text string) error {
	chunks := splitMessage(text, MaxMessageLength)
	for i, chunk := range chunks {
		sendText := chunk
		if i < len(chunks)-1 {
			sendText += "\n...(continued)"
		}
		if err := t.Reply(sendText, telebot.ModeMarkdown); err != nil {
			return err
		}
	}
	return nil
}

// splitMessage 拆分消息
func splitMessage(text string, maxLen int) []string {
	if len(text) <= maxLen {
		return []string{text}
	}

	var chunks []string
	for i := 0; i < len(text); i += maxLen {
		end := i + maxLen
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[i:end])
	}
	return chunks
}
