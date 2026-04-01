package command

import (
	"context"
	"fmt"
	"orange-agent/agent/tools/agent"
	"orange-agent/agent/tools/database"
	"orange-agent/agent/tools/file"
	"orange-agent/agent/tools/git"
	"orange-agent/agent/tools/system"
	"orange-agent/domain"
	"orange-agent/repository"
	"orange-agent/utils/logger"
	"strings"

	"gopkg.in/telebot.v3"
)

// CommandHandler 定义了命令处理器的接口
type CommandHandler interface {
	// Command 返回命令名称（如 "help"）
	Command() string
	// Description 返回命令描述
	Description() string
	// Handle 处理命令
	Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string
}

// CommandManager 管理所有快捷命令
type CommandManager struct {
	handlers map[string]CommandHandler
	log      *logger.Logger
	repo     *repository.Repositories
}

// NewCommandManager 创建新的命令管理器
func NewCommandManager(repo *repository.Repositories) *CommandManager {
	cm := &CommandManager{
		handlers: make(map[string]CommandHandler),
		log:      logger.GetLogger(),
		repo:     repo,
	}

	// 注册所有命令处理器
	cm.registerHandlers()
	return cm
}

// registerHandlers 注册所有命令处理器
func (cm *CommandManager) registerHandlers() {
	// 帮助命令
	cm.Register(&HelpCommand{cm: cm})
	
	// 系统状态命令
	cm.Register(&StatusCommand{})
	
	// 工具列表命令
	cm.Register(&ToolsCommand{})
	
	// 文件操作命令
	cm.Register(&FileListCommand{})
	cm.Register(&FileReadCommand{})
	cm.Register(&FileSearchCommand{})
	
	// Git操作命令
	cm.Register(&GitStatusCommand{})
	cm.Register(&GitCommitCommand{})
	cm.Register(&GitPushCommand{})
	
	// 项目操作命令
	cm.Register(&BuildCommand{})
	cm.Register(&TestCommand{})
	cm.Register(&RebootCommand{})
	cm.Register(&DependencyCheckCommand{})
	cm.Register(&LogViewCommand{})
	
	// 数据库操作命令
	cm.Register(&DbQueryCommand{})
	cm.Register(&DbExecuteCommand{})
	
	// Agent管理命令
	cm.Register(&AgentListCommand{})
	cm.Register(&AgentTestCommand{})
	cm.Register(&AgentAddCommand{})
	cm.Register(&AgentRemoveCommand{})
	cm.Register(&AgentUpdateCommand{})
}

// Register 注册一个命令处理器
func (cm *CommandManager) Register(handler CommandHandler) {
	cm.handlers[strings.ToLower(handler.Command())] = handler
}

// GetCommand 获取指定命令的处理器
func (cm *CommandManager) GetCommand(cmd string) (CommandHandler, bool) {
	handler, exists := cm.handlers[strings.ToLower(cmd)]
	return handler, exists
}

// GetAllCommands 获取所有可用命令
func (cm *CommandManager) GetAllCommands() []CommandHandler {
	commands := make([]CommandHandler, 0, len(cm.handlers))
	for _, handler := range cm.handlers {
		commands = append(commands, handler)
	}
	return commands
}

// Execute 执行命令
func (cm *CommandManager) Execute(ctx context.Context, c telebot.Context, user *domain.User, commandText string) string {
	// 去除斜杠和空格
	commandText = strings.TrimSpace(commandText)
	if !strings.HasPrefix(commandText, "/") {
		return "❌ 命令必须以 '/' 开头"
	}
	
	// 提取命令和参数
	parts := strings.Fields(commandText)
	if len(parts) == 0 {
		return "❌ 命令格式错误"
	}
	
	cmdName := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]
	
	// 查找并执行命令
	handler, exists := cm.GetCommand(cmdName)
	if !exists {
		return fmt.Sprintf("❌ 未知命令: /%s\n📋 使用 /help 查看可用命令", cmdName)
	}
	
	cm.log.Info("执行命令: /%s, 参数: %v", cmdName, args)
	return handler.Handle(ctx, c, user, args)
}

// HelpCommand 帮助命令
type HelpCommand struct {
	cm *CommandManager
}

func (h *HelpCommand) Command() string {
	return "help"
}

func (h *HelpCommand) Description() string {
	return "显示所有可用命令"
}

func (h *HelpCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	commands := h.cm.GetAllCommands()
	
	var response strings.Builder
	response.WriteString("🤖 *Orange Agent 快捷命令*\n\n")
	response.WriteString("以下命令可用于快速执行常见操作：\n\n")
	
	for _, cmd := range commands {
		response.WriteString(fmt.Sprintf("• /%s - %s\n", cmd.Command(), cmd.Description()))
	}
	
	response.WriteString("\n📝 *使用示例：*\n")
	response.WriteString("`/help` - 显示此帮助信息\n")
	response.WriteString("`/status` - 查看系统状态\n")
	response.WriteString("`/list` - 列出项目文件\n")
	response.WriteString("`/read main.go` - 读取文件内容\n")
	response.WriteString("`/git` - 查看Git状态\n")
	response.WriteString("`/commit 修复bug` - 提交更改\n")
	response.WriteString("`/build` - 构建项目\n")
	response.WriteString("`/test` - 运行测试\n")
	response.WriteString("`/agents` - 列出所有Agent\n")
	response.WriteString("`/db SELECT * FROM users` - 执行数据库查询\n")
	
	return response.String()
}

// StatusCommand 系统状态命令
type StatusCommand struct{}

func (s *StatusCommand) Command() string {
	return "status"
}

func (s *StatusCommand) Description() string {
	return "查看系统状态和版本信息"
}

func (s *StatusCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	return "🟢 *系统状态*\n\n• 系统运行正常\n• Orange Agent v1.0.0\n• Telegram Bot 已连接\n• AI Agent 已配置\n\n📊 使用 `/help` 查看所有可用命令"
}

// ToolsCommand 工具列表命令
type ToolsCommand struct{}

func (t *ToolsCommand) Command() string {
	return "tools"
}

func (t *ToolsCommand) Description() string {
	return "列出所有可用工具"
}

func (t *ToolsCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	// 尝试获取工具列表
	var toolList string
	toolList = "🛠️ *可用工具列表*\n\n"
	
	// 添加各类工具
	toolList += "📁 *文件操作:*\n• file_read, file_write, file_delete\n• file_list, file_search, file_rename\n• randomReadFile\n\n"
	
	toolList += "🔧 *系统工具:*\n• build_tools, test_run, project_reboot\n• dependency_check, performance_monitor\n• log_view, env_manage\n\n"
	
	toolList += "🔗 *API工具:*\n• api_tester, web_search\n\n"
	
	toolList += "🗄️ *数据库:*\n• database_query, database_execute\n\n"
	
	toolList += "🤖 *Agent管理:*\n• agent_add, agent_remove, agent_list\n• agent_update, agent_test\n\n"
	
	toolList += "⚙️ *Git操作:*\n• git_commit, git_push, git_diff\n\n"
	
	toolList += "⏰ *时间工具:*\n• curr_time\n\n"
	
	toolList += "📝 使用 `/help` 查看具体命令用法"
	
	return toolList
}