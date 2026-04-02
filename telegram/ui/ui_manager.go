package ui

import (
	"context"
	"fmt"
	"orange-agent/domain"
	"orange-agent/telegram/command"
	"orange-agent/telegram/interfaces"
	"orange-agent/utils/logger"
	"strings"
	"sync"
	"time"

	"github.com/tmc/langchaingo/llms"
	telebot "gopkg.in/telebot.v3"
)

const maxMessageLength = 4096

// UIManager 管理Telegram用户界面交互
type UIManager struct {
	log         *logger.Logger
	cm          *command.CommandManager
	menuManager *MenuManager
	handler     interfaces.MessageHandler
	userStates  map[int64]*UserState
	stateMutex  *sync.RWMutex
}

// UserState 用户状态
type UserState struct {
	LastMenu     string
	LastCommand  string
	LastMessage  string
	LastResponse string
	StateData    map[string]interface{}
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewUIManager 创建新的UI管理器
func NewUIManager(cm *command.CommandManager, handler interfaces.MessageHandler) *UIManager {
	return &UIManager{
		log:         logger.GetLogger(),
		cm:          cm,
		menuManager: NewMenuManager(cm),
		handler:     handler,
		userStates:  make(map[int64]*UserState),
		stateMutex:  &sync.RWMutex{},
	}
}

// GetUserState 获取用户状态
func (um *UIManager) GetUserState(userID int64) *UserState {
	um.stateMutex.RLock()
	state, exists := um.userStates[userID]
	um.stateMutex.RUnlock()

	if !exists {
		state = &UserState{
			StateData: make(map[string]interface{}),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		um.stateMutex.Lock()
		um.userStates[userID] = state
		um.stateMutex.Unlock()
	}

	return state
}

// UpdateUserState 更新用户状态
func (um *UIManager) UpdateUserState(userID int64, state *UserState) {
	um.stateMutex.Lock()
	defer um.stateMutex.Unlock()

	state.UpdatedAt = time.Now()
	um.userStates[userID] = state
}

// HandleMessage 处理消息
func (um *UIManager) HandleMessage(ctx context.Context, c telebot.Context, user *domain.User, text string) (string, *telebot.ReplyMarkup, error) {
	userID := c.Sender().ID
	state := um.GetUserState(userID)
	state.LastMessage = text

	// 检查是否为命令
	if strings.HasPrefix(text, "/") {
		result := um.cm.Execute(ctx, c, user, text)
		state.LastResponse = result
		state.LastCommand = text
		um.UpdateUserState(userID, state)

		menu := um.getMenuForCommand(text)
		return result, menu, nil
	}

	// 检查用户是否在等待参数输入
	if state.StateData["waiting_for_param"] != nil {
		cmd := state.StateData["waiting_for_param"].(string)
		delete(state.StateData, "waiting_for_param")

		fullCmd := fmt.Sprintf("/%s %s", cmd, text)
		result := um.cm.Execute(ctx, c, user, fullCmd)
		state.LastResponse = result
		state.LastCommand = fullCmd
		um.UpdateUserState(userID, state)

		return result, um.menuManager.GetMainMenu(), nil
	}

	// 普通消息，使用AI助手处理
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, text),
	}
	result := um.handler.Handle(ctx, user.ModelName, messages)
	state.LastResponse = result
	um.UpdateUserState(userID, state)

	return result, um.menuManager.GetMainMenu(), nil
}

// HandleCallback 处理按钮回调
func (um *UIManager) HandleCallback(ctx context.Context, c telebot.Context, user *domain.User, data string) (string, *telebot.ReplyMarkup, error) {
	userID := c.Sender().ID
	state := um.GetUserState(userID)

	um.log.Info("处理按钮回调: %s, 用户: %d", data, userID)

	// 处理菜单导航
	if strings.HasPrefix(data, "menu_") {
		menu := um.menuManager.GetMenuByData(data)
		state.LastMenu = data
		um.UpdateUserState(userID, state)

		title := um.getMenuTitle(data)
		return title, menu, nil
	}

	// 处理命令按钮
	if strings.HasPrefix(data, "cmd_") {
		cmd, args := um.menuManager.GetCommandByData(data)

		if len(args) == 0 {
			// 无参数命令直接执行
			fullCmd := fmt.Sprintf("/%s", cmd)
			result := um.cm.Execute(ctx, c, user, fullCmd)
			state.LastResponse = result
			state.LastCommand = fullCmd
			um.UpdateUserState(userID, state)

			return result, um.menuManager.GetMainMenu(), nil
		} else if args[0] == "prompt" {
			// 需要参数的命令
			state.StateData["waiting_for_param"] = cmd
			um.UpdateUserState(userID, state)

			prompt := um.menuManager.GetPromptMessage(cmd)
			return prompt, nil, nil
		} else if cmd == "modelset" && len(args) == 1 {
			// 模型切换命令
			fullCmd := fmt.Sprintf("/%s %s", cmd, args[0])
			result := um.cm.Execute(ctx, c, user, fullCmd)
			state.LastResponse = result
			state.LastCommand = fullCmd
			um.UpdateUserState(userID, state)

			return result, um.menuManager.GetModelMenu(), nil
		}
	}

	// 默认返回主菜单
	return "请选择操作：", um.menuManager.GetMainMenu(), nil
}

// getMenuTitle 获取菜单标题
func (um *UIManager) getMenuTitle(menuData string) string {
	titles := map[string]string{
		"menu_main":    "🤖 *Orange Agent 主菜单*\n\n请选择要执行的操作：",
		"menu_file":    "📁 *文件操作*\n\n选择文件操作：",
		"menu_git":     "🔧 *Git操作*\n\n选择Git操作：",
		"menu_project": "🏗️ *项目管理*\n\n选择项目操作：",
		"menu_db":      "🗄️ *数据库操作*\n\n选择数据库操作：",
		"menu_agent":   "🤖 *Agent管理*\n\n选择Agent操作：",
		"menu_system":  "⚙️ *系统工具*\n\n选择系统操作：",
		"menu_tools":   "🛠️ *系统工具*\n\n选择系统操作：",
		"menu_model":   "🤖 *模型管理*\n\n选择模型操作：",
		"menu_help":    "❓ *帮助中心*\n\n选择帮助内容：",
	}

	if title, ok := titles[menuData]; ok {
		return title
	}
	return "请选择操作："
}

// getMenuForCommand 根据命令获取对应的菜单
func (um *UIManager) getMenuForCommand(cmd string) *telebot.ReplyMarkup {
	cmdPrefix := strings.SplitN(cmd, " ", 2)[0]

	switch {
	case strings.HasPrefix(cmdPrefix, "/list") || strings.HasPrefix(cmdPrefix, "/read") || strings.HasPrefix(cmdPrefix, "/search"):
		return um.menuManager.GetFileMenu()
	case strings.HasPrefix(cmdPrefix, "/git") || strings.HasPrefix(cmdPrefix, "/commit") || strings.HasPrefix(cmdPrefix, "/push"):
		return um.menuManager.GetGitMenu()
	case strings.HasPrefix(cmdPrefix, "/build") || strings.HasPrefix(cmdPrefix, "/test") || strings.HasPrefix(cmdPrefix, "/reboot"):
		return um.menuManager.GetProjectMenu()
	case strings.HasPrefix(cmdPrefix, "/db"):
		return um.menuManager.GetDatabaseMenu()
	case strings.HasPrefix(cmdPrefix, "/agent"):
		return um.menuManager.GetAgentMenu()
	case strings.HasPrefix(cmdPrefix, "/model"):
		return um.menuManager.GetModelMenu()
	default:
		return um.menuManager.GetMainMenu()
	}
}

// GetMainMenu 获取主菜单
func (um *UIManager) GetMainMenu() *telebot.ReplyMarkup {
	return um.menuManager.GetMainMenu()
}

// GetFileMenu 获取文件菜单
func (um *UIManager) GetFileMenu() *telebot.ReplyMarkup {
	return um.menuManager.GetFileMenu()
}

// SendMessageWithMenu 发送带菜单的消息
func (um *UIManager) SendMessageWithMenu(c telebot.Context, text string, menu *telebot.ReplyMarkup) error {
	chunks := splitText(text, maxMessageLength)

	for i, chunk := range chunks {
		var err error
		if i == len(chunks)-1 && menu != nil {
			err = c.Reply(chunk, menu, telebot.ModeMarkdown)
		} else {
			suffix := "\n...(下一部分)"
			if i == len(chunks)-1 {
				suffix = ""
			}
			err = c.Reply(chunk+suffix, telebot.ModeMarkdown)
		}
		if err != nil {
			um.log.Error("发送消息失败: %v", err)
			return err
		}
	}
	return nil
}

// CleanupOldStates 清理过期的用户状态
func (um *UIManager) CleanupOldStates(maxAge time.Duration) {
	um.stateMutex.Lock()
	defer um.stateMutex.Unlock()

	now := time.Now()
	for userID, state := range um.userStates {
		if now.Sub(state.UpdatedAt) > maxAge {
			delete(um.userStates, userID)
		}
	}
}

// splitText 拆分文本
func splitText(text string, maxLen int) []string {
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
