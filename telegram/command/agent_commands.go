package command

import (
	"context"
	"fmt"
	"orange-agent/agent/tools/agent"
	"orange-agent/domain"
	"strings"

	"gopkg.in/telebot.v3"
)

// AgentListCommand Agent列表命令
type AgentListCommand struct{}

func (a *AgentListCommand) Command() string {
	return "agents"
}

func (a *AgentListCommand) Description() string {
	return "列出所有已配置的Agent"
}

func (a *AgentListCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	// 执行Agent列表操作
	result, err := agent.AgentList()
	if err != nil {
		return fmt.Sprintf("❌ 获取Agent列表失败: %v", err)
	}
	
	var response strings.Builder
	response.WriteString("🤖 *已配置的Agent*\n\n")
	
	// 检查是否有Agent
	if strings.TrimSpace(result) == "" {
		response.WriteString("📭 当前没有配置任何Agent")
	} else {
		response.WriteString(result)
	}
	
	return response.String()
}

// AgentTestCommand Agent测试命令
type AgentTestCommand struct{}

func (a *AgentTestCommand) Command() string {
	return "agenttest"
}

func (a *AgentTestCommand) Description() string {
	return "测试指定Agent连接状态"
}

func (a *AgentTestCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	if len(args) == 0 {
		return "❌ 请指定要测试的Agent名称\n📝 用法: `/agenttest <Agent名称>`\n示例: `/agenttest openai`"
	}
	
	agentName := args[0]
	
	// 执行Agent测试操作
	result, err := agent.AgentTest(agentName)
	if err != nil {
		return fmt.Sprintf("❌ 测试Agent失败: %v", err)
	}
	
	return fmt.Sprintf("🔍 *Agent测试结果: %s*\n\n%s", agentName, result)
}

// AgentAddCommand 添加Agent命令
type AgentAddCommand struct{}

func (a *AgentAddCommand) Command() string {
	return "agentadd"
}

func (a *AgentAddCommand) Description() string {
	return "添加新的Agent配置"
}

func (a *AgentAddCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	// 这是一个复杂的操作，需要多个参数
	// 这里只提供使用说明
	response := "🤖 *添加新的Agent*\n\n"
	response += "添加Agent需要多个参数，请使用以下格式：\n\n"
	response += "`/agentadd <名称> <类型> <端点> <API密钥> [模型]`\n\n"
	response += "📋 *参数说明:*\n"
	response += "• *名称*: Agent名称（如：openai、doubao）\n"
	response += "• *类型*: doubao/openai/other\n"
	response += "• *端点*: API端点URL\n"
	response += "• *API密钥*: 您的API密钥\n"
	response += "• *模型*: 可选，默认模型名称\n\n"
	response += "📝 *示例:*\n"
	response += "`/agentadd openai openai https://api.openai.com/v1 sk-xxx gpt-4`\n\n"
	response += "⚠️ 注意：API密钥是敏感信息，请谨慎处理！"
	
	return response
}

// AgentRemoveCommand 删除Agent命令
type AgentRemoveCommand struct{}

func (a *AgentRemoveCommand) Command() string {
	return "agentremove"
}

func (a *AgentRemoveCommand) Description() string {
	return "删除指定Agent配置"
}

func (a *AgentRemoveCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	if len(args) == 0 {
		return "❌ 请指定要删除的Agent名称\n📝 用法: `/agentremove <Agent名称>`\n示例: `/agentremove test-agent`"
	}
	
	agentName := args[0]
	
	// 确认删除
	response := fmt.Sprintf("⚠️ *确认删除Agent*\n\n您将要删除 Agent: **%s**\n\n此操作不可撤销！\n\n回复 `确认删除` 来继续。", agentName)
	
	return response
}

// AgentUpdateCommand 更新Agent命令
type AgentUpdateCommand struct{}

func (a *AgentUpdateCommand) Command() string {
	return "agentupdate"
}

func (a *AgentUpdateCommand) Description() string {
	return "更新Agent配置信息"
}

func (a *AgentUpdateCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	response := "🤖 *更新Agent配置*\n\n"
	response += "更新Agent配置，请使用以下格式：\n\n"
	response += "`/agentupdate <名称> [类型] [端点] [API密钥] [模型]`\n\n"
	response += "📋 *参数说明:*\n"
	response += "• *名称*: 必须，要更新的Agent名称\n"
	response += "• *类型*: 可选，doubao/openai/other\n"
	response += "• *端点*: 可选，API端点URL\n"
	response += "• *API密钥*: 可选，新的API密钥\n"
	response += "• *模型*: 可选，新的模型名称\n\n"
	response += "📝 *示例:*\n"
	response += "`/agentupdate openai https://new-api.openai.com/v1`\n"
	response += "`/agentupdate openai gpt-4-turbo`\n\n"
	response += "⚠️ 注意：至少提供一个更新参数！"
	
	return response
}