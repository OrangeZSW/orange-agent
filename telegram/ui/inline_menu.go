package ui

import (
	telebot "gopkg.in/telebot.v3"
)

// InlineMenuManager 管理内联菜单（在消息中显示）
type InlineMenuManager struct {
	menuManager *MenuManager
}

// NewInlineMenuManager 创建新的内联菜单管理器
func NewInlineMenuManager(mm *MenuManager) *InlineMenuManager {
	return &InlineMenuManager{
		menuManager: mm,
	}
}

// GetQuickActionsMenu 获取快速操作内联菜单
func (im *InlineMenuManager) GetQuickActionsMenu() *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{}

	// 第一行：常用操作
	menu.Row(
		menu.Data("📋 文件列表", "inline_list"),
		menu.Data("📊 Git状态", "inline_git"),
		menu.Data("🔨 构建", "inline_build"),
	)

	// 第二行：系统操作
	menu.Row(
		menu.Data("📊 状态", "inline_status"),
		menu.Data("🛠️ 工具", "inline_tools"),
		menu.Data("🤖 模型", "inline_model"),
	)

	// 第三行：帮助
	menu.Row(
		menu.Data("❓ 帮助", "inline_help"),
		menu.Data("📖 命令", "inline_commands"),
	)

	return menu
}

// GetFileInlineMenu 获取文件操作内联菜单
func (im *InlineMenuManager) GetFileInlineMenu() *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{}

	// 第一行：查看操作
	menu.Row(
		menu.Data("📁 列出", "inline_list"),
		menu.Data("📄 读取", "inline_read_prompt"),
		menu.Data("🔍 搜索", "inline_search_prompt"),
	)

	// 第二行：编辑操作
	menu.Row(
		menu.Data("📝 写入", "inline_write_prompt"),
		menu.Data("✂️ 重命名", "inline_rename_prompt"),
		menu.Data("🗑️ 删除", "inline_delete_prompt"),
	)

	// 第三行：导航
	menu.Row(
		menu.Data("⬅️ 返回", "inline_back"),
		menu.Data("🏠 主菜单", "inline_main"),
	)

	return menu
}

// GetGitInlineMenu 获取Git操作内联菜单
func (im *InlineMenuManager) GetGitInlineMenu() *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{}

	// 第一行：状态和提交
	menu.Row(
		menu.Data("📊 状态", "inline_git"),
		menu.Data("📥 差异", "inline_diff"),
		menu.Data("💾 提交", "inline_commit_prompt"),
	)

	// 第二行：推送和分支
	menu.Row(
		menu.Data("📤 推送", "inline_push_prompt"),
		menu.Data("🌿 分支", "inline_branch_prompt"),
		menu.Data("🔄 拉取", "inline_pull_prompt"),
	)

	// 第三行：导航
	menu.Row(
		menu.Data("⬅️ 返回", "inline_back"),
		menu.Data("🏠 主菜单", "inline_main"),
	)

	return menu
}

// GetProjectInlineMenu 获取项目管理内联菜单
func (im *InlineMenuManager) GetProjectInlineMenu() *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{}

	// 第一行：构建和测试
	menu.Row(
		menu.Data("🔨 构建", "inline_build"),
		menu.Data("🧪 测试", "inline_test"),
		menu.Data("🔄 重启", "inline_reboot"),
	)

	// 第二行：依赖和日志
	menu.Row(
		menu.Data("📦 依赖", "inline_deps"),
		menu.Data("📋 日志", "inline_logs"),
		menu.Data("📝 环境", "inline_env_prompt"),
	)

	// 第三行：监控和搜索
	menu.Row(
		menu.Data("📈 监控", "inline_perf_prompt"),
		menu.Data("🌐 搜索", "inline_websearch_prompt"),
		menu.Data("🔗 API测试", "inline_api_prompt"),
	)

	// 第四行：导航
	menu.Row(
		menu.Data("⬅️ 返回", "inline_back"),
		menu.Data("🏠 主菜单", "inline_main"),
	)

	return menu
}

// GetDatabaseInlineMenu 获取数据库内联菜单
func (im *InlineMenuManager) GetDatabaseInlineMenu() *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{}

	// 第一行：查询操作
	menu.Row(
		menu.Data("🔍 查询", "inline_db_prompt"),
		menu.Data("📊 统计", "inline_stats_prompt"),
		menu.Data("📋 表结构", "inline_schema_prompt"),
	)

	// 第二行：写操作
	menu.Row(
		menu.Data("✏️ 插入", "inline_insert_prompt"),
		menu.Data("🔄 更新", "inline_update_prompt"),
		menu.Data("🗑️ 删除", "inline_delete_db_prompt"),
	)

	// 第三行：导航
	menu.Row(
		menu.Data("⬅️ 返回", "inline_back"),
		menu.Data("🏠 主菜单", "inline_main"),
	)

	return menu
}

// GetAgentInlineMenu 获取Agent管理内联菜单
func (im *InlineMenuManager) GetAgentInlineMenu() *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{}

	// 第一行：查看和测试
	menu.Row(
		menu.Data("📋 列表", "inline_agents"),
		menu.Data("🧪 测试", "inline_agenttest_prompt"),
		menu.Data("📊 状态", "inline_agentstatus"),
	)

	// 第二行：管理操作
	menu.Row(
		menu.Data("➕ 添加", "inline_agentadd_prompt"),
		menu.Data("✖️ 删除", "inline_agentremove_prompt"),
		menu.Data("🔄 更新", "inline_agentupdate_prompt"),
	)

	// 第三行：导航
	menu.Row(
		menu.Data("⬅️ 返回", "inline_back"),
		menu.Data("🏠 主菜单", "inline_main"),
	)

	return menu
}

// GetModelInlineMenu 获取模型管理内联菜单
func (im *InlineMenuManager) GetModelInlineMenu() *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{}

	// 第一行：查看和切换
	menu.Row(
		menu.Data("📋 列表", "inline_model"),
		menu.Data("🔄 切换", "inline_modelset_prompt"),
		menu.Data("⚡ 快速切换", "inline_quickmodel"),
	)

	// 第二行：常用模型
	menu.Row(
		menu.Data("🤖 GPT-4", "inline_model_gpt4"),
		menu.Data("🧠 Claude", "inline_model_claude"),
		menu.Data("💎 Gemini", "inline_model_gemini"),
	)

	// 第三行：导航
	menu.Row(
		menu.Data("⬅️ 返回", "inline_back"),
		menu.Data("🏠 主菜单", "inline_main"),
	)

	return menu
}

// GetHelpInlineMenu 获取帮助内联菜单
func (im *InlineMenuManager) GetHelpInlineMenu() *telebot.ReplyMarkup {
	menu := &telebot.ReplyMarkup{}

	// 第一行：帮助内容
	menu.Row(
		menu.Data("📖 帮助", "inline_help"),
		menu.Data("📋 命令", "inline_commands"),
		menu.Data("🛠️ 示例", "inline_examples"),
	)

	// 第二行：指南和技巧
	menu.Row(
		menu.Data("📝 指南", "inline_guide"),
		menu.Data("💡 技巧", "inline_tips"),
		menu.Data("❓ 常见问题", "inline_faq"),
	)

	// 第三行：导航
	menu.Row(
		menu.Data("⬅️ 返回", "inline_back"),
		menu.Data("🏠 主菜单", "inline_main"),
	)

	return menu
}

// GetContextualMenu 根据上下文获取内联菜单
func (im *InlineMenuManager) GetContextualMenu(contextType string) *telebot.ReplyMarkup {
	switch contextType {
	case "file":
		return im.GetFileInlineMenu()
	case "git":
		return im.GetGitInlineMenu()
	case "project":
		return im.GetProjectInlineMenu()
	case "database":
		return im.GetDatabaseInlineMenu()
	case "agent":
		return im.GetAgentInlineMenu()
	case "model":
		return im.GetModelInlineMenu()
	case "help":
		return im.GetHelpInlineMenu()
	default:
		return im.GetQuickActionsMenu()
	}
}
