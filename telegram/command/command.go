package command

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/agent/tools"
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

	// 模型切换命令
	cm.Register(&ModelCommand{})
	cm.Register(&ModelSetCommand{})
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
	response.WriteString("`/model` - 查看和切换AI模型\n")
	response.WriteString("`/modelset gpt-4` - 快速切换到指定模型\n")

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
	// 显示当前模型信息
	currentModel := "未设置"
	if user != nil && user.ModelName != "" {
		currentModel = user.ModelName
	}

	return fmt.Sprintf("🟢 *系统状态*\n\n• 系统运行正常\n• Orange Agent v1.0.0\n• Telegram Bot 已连接\n• AI Agent 已配置\n• *当前模型:* %s\n\n📊 使用 `/help` 查看所有可用命令", currentModel)
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
	// 获取所有工具
	allTools := tools.GetTools()

	var response strings.Builder
	response.WriteString("🛠️ *可用工具列表*\n\n")
	response.WriteString(fmt.Sprintf("📊 共 %d 个工具\n\n", len(allTools)))

	// 按类别分组显示
	response.WriteString("📁 *文件操作:*\n")
	for name, tool := range allTools {
		if strings.Contains(name, "file") {
			response.WriteString(fmt.Sprintf("• %s - %s\n", name, tool.Description))
		}
	}

	response.WriteString("\n🔧 *系统工具:*\n")
	for name, tool := range allTools {
		if strings.Contains(name, "build") || strings.Contains(name, "test") || strings.Contains(name, "depend") ||
			strings.Contains(name, "log") || strings.Contains(name, "env") || strings.Contains(name, "perform") ||
			strings.Contains(name, "reboot") {
			response.WriteString(fmt.Sprintf("• %s - %s\n", name, tool.Description))
		}
	}

	response.WriteString("\n🔗 *API工具:*\n")
	for name, tool := range allTools {
		if strings.Contains(name, "api") || strings.Contains(name, "web") {
			response.WriteString(fmt.Sprintf("• %s - %s\n", name, tool.Description))
		}
	}

	response.WriteString("\n🗄️ *数据库:*\n")
	for name, tool := range allTools {
		if strings.Contains(name, "database") {
			response.WriteString(fmt.Sprintf("• %s - %s\n", name, tool.Description))
		}
	}

	response.WriteString("\n🤖 *Agent管理:*\n")
	for name, tool := range allTools {
		if strings.Contains(name, "agent") {
			response.WriteString(fmt.Sprintf("• %s - %s\n", name, tool.Description))
		}
	}

	response.WriteString("\n⚙️ *Git操作:*\n")
	for name, tool := range allTools {
		if strings.Contains(name, "git") {
			response.WriteString(fmt.Sprintf("• %s - %s\n", name, tool.Description))
		}
	}

	response.WriteString("\n⏰ *时间工具:*\n")
	for name, tool := range allTools {
		if strings.Contains(name, "time") {
			response.WriteString(fmt.Sprintf("• %s - %s\n", name, tool.Description))
		}
	}

	response.WriteString("\n📝 使用 `/help` 查看具体命令用法")

	return response.String()
}

// executeTool 执行工具函数
func executeTool(toolName string, params interface{}) (string, error) {
	allTools := tools.GetTools()
	tool, exists := allTools[toolName]
	if !exists {
		return "", fmt.Errorf("工具 %s 不存在", toolName)
	}

	// 将参数转换为JSON字符串
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("参数转换失败: %v", err)
	}

	// 调用工具
	result, err := tool.Call(context.Background(), string(paramsJSON))
	if err != nil {
		return "", fmt.Errorf("工具执行失败: %v", err)
	}

	return result, nil
}

// executeDBTool 执行数据库工具函数
func executeDBTool(query string, args []interface{}) (string, error) {
	params := map[string]interface{}{
		"query": query,
		"args":  args,
	}

	// 判断是查询还是执行操作
	toolName := "database_query"
	if strings.HasPrefix(strings.ToUpper(query), "INSERT") ||
		strings.HasPrefix(strings.ToUpper(query), "UPDATE") ||
		strings.HasPrefix(strings.ToUpper(query), "DELETE") {
		toolName = "database_execute"
	}

	return executeTool(toolName, params)
}