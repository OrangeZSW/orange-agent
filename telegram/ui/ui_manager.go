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

// UIManager 管理Telegram用户界面交互
type UIManager struct {
	log           *logger.Logger
	cm            *command.CommandManager
	menuManager   *MenuManager
	answer        interfaces.Ansewer
	userStates    map[int64]*UserState
	stateMutex    *sync.RWMutex
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
func NewUIManager(cm *command.CommandManager, answer interfaces.Ansewer) *UIManager {
	return &UIManager{
		log:         logger.GetLogger(),
		cm:          cm,
		menuManager: NewMenuManager(cm),
		answer:      answer,
		userStates:  make(map[int64]*UserState),
		stateMutex:  &sync.RWMutex{},
	}
}

// GetUserState 获取用户状态
func (um *UIManager) GetUserState(userID int64) *UserState {
	um.stateMutex.RLock()
	defer um.stateMutex.RUnlock()
	
	state, exists := um.userStates[userID]
	if !exists {
		state = &UserState{
			StateData: make(map[string]interface{}),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
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
	
	// 检查是否为按钮回调
	if c.Callback() != nil && c.Data() != "" {
		return um.HandleCallback(ctx, c, user, c.Data())
	}
	
	// 检查是否为命令
	if strings.HasPrefix(text, "/") {
		// 执行命令
		result := um.cm.Execute(ctx, c, user, text)
		state.LastResponse = result
		state.LastCommand = text
		um.UpdateUserState(userID, state)
		
		// 根据命令类型决定是否显示菜单
		menu := um.getMenuForCommand(text)
		return result, menu, nil
	}
	
	// 检查用户是否在等待参数输入
	if state.StateData["waiting_for_param"] != nil {
		cmd := state.StateData["waiting_for_param"].(string)
		delete(state.StateData, "waiting_for_param")
		
		// 执行带参数的命令
		fullCmd := fmt.Sprintf("/%s %s", cmd, text)
		result := um.cm.Execute(ctx, c, user, fullCmd)
		state.LastResponse = result
		state.LastCommand = fullCmd
		um.UpdateUserState(userID, state)
		
		// 返回主菜单
		return result, um.menuManager.GetMainMenu(), nil
	}
	
	// 普通消息，使用AI助手处理 - 转换为 llms.MessageContent
	messageContent := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, text),
	}
	result := um.answer.TeleGramChat(ctx, user.ModelName, messageContent)
	state.LastResponse = result
	um.UpdateUserState(userID, state)
	
	// 显示主菜单
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
		
		// 返回菜单标题和对应的键盘
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
			
			// 返回结果和主菜单
			return result, um.menuManager.GetMainMenu(), nil
		} else if args[0] == "prompt" {
			// 需要参数的命令
			state.StateData["waiting_for_param"] = cmd
			um.UpdateUserState(userID, state)
			
			// 返回参数提示
			prompt := um.menuManager.GetPromptMessage(cmd)
			return prompt, nil, nil
		} else if cmd == "modelset" && len(args) == 1 {
			// 模型切换命令，直接执行
			fullCmd := fmt.Sprintf("/%s %s", cmd, args[0])
			result := um.cm.Execute(ctx, c, user, fullCmd)
			state.LastResponse = result
			state.LastCommand = fullCmd
			um.UpdateUserState(userID, state)
			
			// 返回切换结果和模型菜单
			return result, um.menuManager.GetModelMenu(), nil
		}
	}
	
	// 默认返回主菜单
	return "请选择操作：", um.menuManager.GetMainMenu(), nil
}

// getMenuTitle 获取菜单标题
func (um *UIManager) getMenuTitle(menuData string) string {
	switch menuData {
	case "menu_main":
		return "🤖 *Orange Agent 主菜单*\n\n请选择要执行的操作："
	case "menu_file":
		return "📁 *文件操作*\n\n选择文件操作："
	case "menu_git":
		return "🔧 *Git操作*\n\n选择Git操作："
	case "menu_project":
		return "🏗️ *项目管理*\n\n选择项目操作："
	case "menu_db":
		return "🗄️ *数据库操作*\n\n选择数据库操作："
	case "menu_agent":
		return "🤖 *Agent管理*\n\n选择Agent操作："
	case "menu_system":
		return "⚙️ *系统工具*\n\n选择系统操作："
	case "menu_tools":
		return "🛠️ *系统工具*\n\n选择系统操作："
	case "menu_model":
		return "🤖 *模型管理*\n\n选择模型操作："
	case "menu_help":
		return "❓ *帮助中心*\n\n选择帮助内容："
	default:
		return "请选择操作："
	}
}

// getMenuForCommand 根据命令获取对应的菜单
func (um *UIManager) getMenuForCommand(cmd string) *telebot.ReplyMarkup {
	// 根据命令类型返回相应的菜单
	if strings.Contains(cmd, "/list") || strings.Contains(cmd, "/read") || strings.Contains(cmd, "/search") {
		return um.menuManager.GetFileMenu()
	} else if strings.Contains(cmd, "/git") || strings.Contains(cmd, "/commit") || strings.Contains(cmd, "/push") {
		return um.menuManager.GetGitMenu()
	} else if strings.Contains(cmd, "/build") || strings.Contains(cmd, "/test") || strings.Contains(cmd, "/reboot") {
		return um.menuManager.GetProjectMenu()
	} else if strings.Contains(cmd, "/db") {
		return um.menuManager.GetDatabaseMenu()
	} else if strings.Contains(cmd, "/agent") {
		return um.menuManager.GetAgentMenu()
	} else if strings.Contains(cmd, "/model") {
		return um.menuManager.GetModelMenu()
	} else if strings.Contains(cmd, "/help") || strings.Contains(cmd, "/status") || strings.Contains(cmd, "/tools") {
		return um.menuManager.GetMainMenu()
	}
	
	return um.menuManager.GetMainMenu()
}

// SendMessageWithMenu 发送带菜单的消息
func (um *UIManager) SendMessageWithMenu(c telebot.Context, text string, menu *telebot.ReplyMarkup) error {
	if menu != nil {
		return c.Reply(text, menu, telebot.ModeMarkdown)
	}
	return c.Reply(text, telebot.ModeMarkdown)
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

// 添加公共方法以暴露菜单管理器的方法
func (um *UIManager) GetMenuManager() *telebot.ReplyMarkup {
	return um.menuManager.GetMainMenu()
}

func (um *UIManager) GetFileMenu() *telebot.ReplyMarkup {
	return um.menuManager.GetFileMenu()
}

func (um *UIManager) GetGitMenu() *telebot.ReplyMarkup {
	return um.menuManager.GetGitMenu()
}

func (um *UIManager) GetProjectMenu() *telebot.ReplyMarkup {
	return um.menuManager.GetProjectMenu()
}

func (um *UIManager) GetDatabaseMenu() *telebot.ReplyMarkup {
	return um.menuManager.GetDatabaseMenu()
}

func (um *UIManager) GetAgentMenu() *telebot.ReplyMarkup {
	return um.menuManager.GetAgentMenu()
}

func (um *UIManager) GetSystemMenu() *telebot.ReplyMarkup {
	return um.menuManager.GetSystemMenu()
}

func (um *UIManager) GetModelMenu() *telebot.ReplyMarkup {
	return um.menuManager.GetModelMenu()
}

func (um *UIManager) GetHelpMenu() *telebot.ReplyMarkup {
	return um.menuManager.GetHelpMenu()
}

func (um *UIManager) GetToolsMenu() *telebot.ReplyMarkup {
	return um.menuManager.GetSystemMenu()
}