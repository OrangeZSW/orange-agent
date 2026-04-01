package ui

import (
	"encoding/json"
	"orange-agent/domain"
	"orange-agent/telegram/command"
	"orange-agent/utils/logger"
	"strings"

	telebot "gopkg.in/telebot.v3"
)

// MenuManager 管理Telegram按钮菜单
type MenuManager struct {
	log       *logger.Logger
	cmds      *command.CommandManager
	keyboards map[string]*telebot.ReplyMarkup
}

// NewMenuManager 创建新的菜单管理器
func NewMenuManager(cmds *command.CommandManager) *MenuManager {
	return &MenuManager{
		log:       logger.GetLogger(),
		cmds:      cmds,
		keyboards: make(map[string]*telebot.ReplyMarkup),
	}
}

// GetMainMenu 获取主菜单键盘
func (mm *MenuManager) GetMainMenu() *telebot.ReplyMarkup {
	keyboard := &telebot.ReplyMarkup{}
	keyboard.Inline(
		// 第一行：核心功能
		keyboard.Row(
			telebot.Btn{Text: "📁 文件", Data: "menu_file"},
			telebot.Btn{Text: "🔧 Git", Data: "menu_git"},
			telebot.Btn{Text: "🏗️ 项目", Data: "menu_project"},
		),
		// 第二行：数据库和Agent
		keyboard.Row(
			telebot.Btn{Text: "🗄️ 数据库", Data: "menu_db"},
			telebot.Btn{Text: "🤖 Agent", Data: "menu_agent"},
			telebot.Btn{Text: "⚙️ 系统", Data: "menu_system"},
		),
		// 第三行：工具和模型
		keyboard.Row(
			telebot.Btn{Text: "🛠️ 工具", Data: "menu_tools"},
			telebot.Btn{Text: "🤖 模型", Data: "menu_model"},
			telebot.Btn{Text: "❓ 帮助", Data: "menu_help"},
		),
	)
	return keyboard
}

// GetFileMenu 获取文件操作菜单
func (mm *MenuManager) GetFileMenu() *telebot.ReplyMarkup {
	keyboard := &telebot.ReplyMarkup{}
	keyboard.Inline(
		keyboard.Row(
			telebot.Btn{Text: "📋 列出文件", Data: "cmd_list"},
			telebot.Btn{Text: "📄 读取文件", Data: "cmd_read_prompt"},
		),
		keyboard.Row(
			telebot.Btn{Text: "🔍 搜索文件", Data: "cmd_search_prompt"},
			telebot.Btn{Text: "📝 写入文件", Data: "cmd_write_prompt"},
		),
		keyboard.Row(
			telebot.Btn{Text: "⬅️ 返回主菜单", Data: "menu_main"},
		),
	)
	return keyboard
}

// GetGitMenu 获取Git操作菜单
func (mm *MenuManager) GetGitMenu() *telebot.ReplyMarkup {
	keyboard := &telebot.ReplyMarkup{}
	keyboard.Inline(
		keyboard.Row(
			telebot.Btn{Text: "📊 Git状态", Data: "cmd_git"},
			telebot.Btn{Text: "💾 提交更改", Data: "cmd_commit_prompt"},
		),
		keyboard.Row(
			telebot.Btn{Text: "📤 推送代码", Data: "cmd_push_prompt"},
			telebot.Btn{Text: "📥 差异对比", Data: "cmd_diff_prompt"},
		),
		keyboard.Row(
			telebot.Btn{Text: "⬅️ 返回主菜单", Data: "menu_main"},
		),
	)
	return keyboard
}

// GetProjectMenu 获取项目管理菜单
func (mm *MenuManager) GetProjectMenu() *telebot.ReplyMarkup {
	keyboard := &telebot.ReplyMarkup{}
	keyboard.Inline(
		keyboard.Row(
			telebot.Btn{Text: "🔨 构建项目", Data: "cmd_build"},
			telebot.Btn{Text: "🧪 运行测试", Data: "cmd_test"},
		),
		keyboard.Row(
			telebot.Btn{Text: "🔄 重启项目", Data: "cmd_reboot"},
			telebot.Btn{Text: "📦 检查依赖", Data: "cmd_deps"},
		),
		keyboard.Row(
			telebot.Btn{Text: "📋 查看日志", Data: "cmd_logs"},
			telebot.Btn{Text: "📝 环境变量", Data: "cmd_env_prompt"},
		),
		keyboard.Row(
			telebot.Btn{Text: "⬅️ 返回主菜单", Data: "menu_main"},
		),
	)
	return keyboard
}

// GetDatabaseMenu 获取数据库菜单
func (mm *MenuManager) GetDatabaseMenu() *telebot.ReplyMarkup {
	keyboard := &telebot.ReplyMarkup{}
	keyboard.Inline(
		keyboard.Row(
			telebot.Btn{Text: "🔍 查询数据", Data: "cmd_db_prompt"},
			telebot.Btn{Text: "✏️ 执行SQL", Data: "cmd_dbe_prompt"},
		),
		keyboard.Row(
			telebot.Btn{Text: "⬅️ 返回主菜单", Data: "menu_main"},
		),
	)
	return keyboard
}

// GetAgentMenu 获取Agent管理菜单
func (mm *MenuManager) GetAgentMenu() *telebot.ReplyMarkup {
	keyboard := &telebot.ReplyMarkup{}
	keyboard.Inline(
		keyboard.Row(
			telebot.Btn{Text: "📋 列出Agent", Data: "cmd_agents"},
			telebot.Btn{Text: "🧪 测试Agent", Data: "cmd_agenttest_prompt"},
		),
		keyboard.Row(
			telebot.Btn{Text: "➕ 添加Agent", Data: "cmd_agentadd_prompt"},
			telebot.Btn{Text: "✖️ 删除Agent", Data: "cmd_agentremove_prompt"},
		),
		keyboard.Row(
			telebot.Btn{Text: "🔄 更新Agent", Data: "cmd_agentupdate_prompt"},
			telebot.Btn{Text: "⬅️ 返回主菜单", Data: "menu_main"},
		),
	)
	return keyboard
}

// GetSystemMenu 获取系统菜单
func (mm *MenuManager) GetSystemMenu() *telebot.ReplyMarkup {
	keyboard := &telebot.ReplyMarkup{}
	keyboard.Inline(
		keyboard.Row(
			telebot.Btn{Text: "📊 系统状态", Data: "cmd_status"},
			telebot.Btn{Text: "🛠️ 工具列表", Data: "cmd_tools"},
		),
		keyboard.Row(
			telebot.Btn{Text: "⏰ 当前时间", Data: "cmd_time"},
			telebot.Btn{Text: "📈 性能监控", Data: "cmd_perf_prompt"},
		),
		keyboard.Row(
			telebot.Btn{Text: "🌐 Web搜索", Data: "cmd_websearch_prompt"},
			telebot.Btn{Text: "🔗 API测试", Data: "cmd_api_prompt"},
		),
		keyboard.Row(
			telebot.Btn{Text: "⬅️ 返回主菜单", Data: "menu_main"},
		),
	)
	return keyboard
}

// GetModelMenu 获取模型菜单
func (mm *MenuManager) GetModelMenu() *telebot.ReplyMarkup {
	keyboard := &telebot.ReplyMarkup{}
	
	// 首先获取所有可用模型
	var rows []telebot.Row
	
	// 第一行：查看模型按钮
	rows = append(rows, keyboard.Row(
		telebot.Btn{Text: "📋 查看当前模型", Data: "cmd_model"},
	))
	
	// 获取所有Agent配置
	agentConfigs, err := mm.getAgentConfigs()
	if err == nil && len(agentConfigs) > 0 {
		// 收集所有模型
		allModels := make(map[string]bool)
		for _, agent := range agentConfigs {
			for _, model := range agent.Models {
				allModels[model] = true
			}
		}
		
		if len(allModels) > 0 {
			// 添加模型切换标题
			rows = append(rows, keyboard.Row(
				telebot.Btn{Text: "🔄 点击切换模型:", Data: "menu_model"},
			))
			
			// 为每个模型创建按钮，每行最多2个
			var currentRow []telebot.Btn
			for modelName := range allModels {
				btn := telebot.Btn{
					Text: modelName,
					Data: "cmd_modelset_" + modelName,
				}
				currentRow = append(currentRow, btn)
				
				if len(currentRow) == 2 {
					rows = append(rows, keyboard.Row(currentRow...))
					currentRow = []telebot.Btn{}
				}
			}
			
			// 添加剩余的按钮
			if len(currentRow) > 0 {
				rows = append(rows, keyboard.Row(currentRow...))
			}
		}
	}
	
	// 最后一行：返回主菜单
	rows = append(rows, keyboard.Row(
		telebot.Btn{Text: "⬅️ 返回主菜单", Data: "menu_main"},
	))
	
	keyboard.Inline(rows...)
	return keyboard
}

// getAgentConfigs 获取所有Agent配置
func (mm *MenuManager) getAgentConfigs() ([]domain.AgentConfig, error) {
	// 使用工具调用获取Agent配置
	result, err := command.ExecuteTool("agent_list", map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	
	// 解析JSON结果
	if strings.TrimSpace(result) == "" {
		return []domain.AgentConfig{}, nil
	}
	
	// 首先解析为包含data字段的对象
	var responseObj map[string]interface{}
	if err := json.Unmarshal([]byte(result), &responseObj); err != nil {
		return nil, err
	}
	
	// 从data字段获取Agent列表
	dataField, ok := responseObj["data"]
	if !ok {
		return []domain.AgentConfig{}, nil
	}
	
	// 将data字段转换为JSON字符串再解析为Agent数组
	dataJSON, err := json.Marshal(dataField)
	if err != nil {
		return nil, err
	}
	
	var agentConfigs []domain.AgentConfig
	if err := json.Unmarshal(dataJSON, &agentConfigs); err != nil {
		return nil, err
	}
	
	return agentConfigs, nil
}

// GetHelpMenu 获取帮助菜单
func (mm *MenuManager) GetHelpMenu() *telebot.ReplyMarkup {
	keyboard := &telebot.ReplyMarkup{}
	keyboard.Inline(
		keyboard.Row(
			telebot.Btn{Text: "📖 命令帮助", Data: "cmd_help"},
			telebot.Btn{Text: "📋 快速指南", Data: "cmd_quickstart"},
		),
		keyboard.Row(
			telebot.Btn{Text: "🛠️ 使用示例", Data: "cmd_examples"},
			telebot.Btn{Text: "📝 使用技巧", Data: "cmd_tips"},
		),
		keyboard.Row(
			telebot.Btn{Text: "⬅️ 返回主菜单", Data: "menu_main"},
		),
	)
	return keyboard
}

// GetMenuByData 根据数据获取菜单
func (mm *MenuManager) GetMenuByData(data string) *telebot.ReplyMarkup {
	switch data {
	case "menu_main":
		return mm.GetMainMenu()
	case "menu_file":
		return mm.GetFileMenu()
	case "menu_git":
		return mm.GetGitMenu()
	case "menu_project":
		return mm.GetProjectMenu()
	case "menu_db":
		return mm.GetDatabaseMenu()
	case "menu_agent":
		return mm.GetAgentMenu()
	case "menu_system":
		return mm.GetSystemMenu()
	case "menu_tools":
		return mm.GetSystemMenu() // tools和system共用
	case "menu_model":
		return mm.GetModelMenu()
	case "menu_help":
		return mm.GetHelpMenu()
	default:
		return mm.GetMainMenu()
	}
}

// GetCommandByData 根据按钮数据获取对应的命令
func (mm *MenuManager) GetCommandByData(data string) (string, []string) {
	// 处理菜单导航
	if strings.HasPrefix(data, "menu_") {
		return "", nil
	}

	// 处理简单命令
	if strings.HasPrefix(data, "cmd_") {
		cmdName := strings.TrimPrefix(data, "cmd_")

		// 特殊处理模型切换命令：cmd_modelset_<模型名>
		if strings.HasPrefix(cmdName, "modelset_") {
			modelName := strings.TrimPrefix(cmdName, "modelset_")
			return "modelset", []string{modelName}
		}

		// 特殊处理需要参数的命令
		if strings.HasSuffix(cmdName, "_prompt") {
			baseCmd := strings.TrimSuffix(cmdName, "_prompt")
			return baseCmd, []string{"prompt"}
		}

		return cmdName, nil
	}

	return "", nil
}

// GetPromptMessage 获取参数提示消息
func (mm *MenuManager) GetPromptMessage(cmd string) string {
	switch cmd {
	case "read":
		return "📄 请输入要读取的文件路径（例如：main.go）："
	case "search":
		return "🔍 请输入要搜索的内容："
	case "write":
		return "📝 请输入文件路径和内容（格式：路径|内容）："
	case "commit":
		return "💾 请输入提交信息："
	case "push":
		return "📤 请输入分支名称（可选，默认当前分支）："
	case "diff":
		return "📥 请输入要对比的文件（可选）："
	case "env":
		return "📝 请输入环境变量操作（格式：get/set/list 键 值）："
	case "db":
		return "🔍 请输入SQL查询语句："
	case "dbe":
		return "✏️ 请输入SQL执行语句（请确认操作安全）："
	case "agenttest":
		return "🧪 请输入要测试的Agent名称："
	case "agentadd":
		return "➕ 请输入Agent配置（格式：名称|端点|密钥|类型|模型）："
	case "agentremove":
		return "✖️ 请输入要删除的Agent名称："
	case "agentupdate":
		return "🔄 请输入更新信息（格式：名称|字段|值）："
	case "modelset":
		return "🔄 请输入要切换的模型名称："
	case "perf":
		return "📈 请输入监控指标（cpu/memory/disk/all）："
	case "websearch":
		return "🌐 请输入搜索关键词："
	case "api":
		return "🔗 请输入API URL（格式：URL|方法|数据）："
	default:
		return "请输入参数："
	}
}