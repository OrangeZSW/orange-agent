package command

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/domain"
	"strings"

	"gopkg.in/telebot.v3"
)

// ModelCommand 模型切换命令
type ModelCommand struct{}

func (m *ModelCommand) Command() string {
	return "model"
}

func (m *ModelCommand) Description() string {
	return "切换当前使用的AI模型"
}

func (m *ModelCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	if len(args) == 0 {
		// 显示当前模型和可用模型列表
		return m.showModels(ctx, c, user, args)
	}
	
	// 切换模型
	return m.switchModel(ctx, c, user, args[0])
}

// showModels 显示当前模型和可用模型列表
func (m *ModelCommand) showModels(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	var response strings.Builder
	
	// 获取当前用户的模型
	currentModel := "未设置"
	if user != nil && user.ModelName != "" {
		currentModel = user.ModelName
	}
	
	// 获取所有Agent配置以显示可用模型
	agentConfigs, err := getAgentConfigs()
	if err != nil {
		return fmt.Sprintf("❌ 获取可用模型失败: %v", err)
	}
	
	response.WriteString("🤖 *模型切换*\n\n")
	response.WriteString(fmt.Sprintf("📋 *当前模型:* %s\n\n", currentModel))
	response.WriteString("🔄 *可用模型列表:*\n")
	
	// 收集所有模型
	allModels := make(map[string]string) // modelName -> agentName
	for _, agent := range agentConfigs {
		for _, model := range agent.Models {
			allModels[model] = agent.Name
		}
	}
	
	if len(allModels) == 0 {
		response.WriteString("📭 当前没有可用的模型\n")
		response.WriteString("请先使用 `/agentadd` 命令添加Agent配置\n")
	} else {
		// 按字母顺序显示模型
		modelNames := make([]string, 0, len(allModels))
		for modelName := range allModels {
			modelNames = append(modelNames, modelName)
		}
		
		// 简单排序
		for i := 0; i < len(modelNames)-1; i++ {
			for j := i + 1; j < len(modelNames); j++ {
				if modelNames[i] > modelNames[j] {
					modelNames[i], modelNames[j] = modelNames[j], modelNames[i]
				}
			}
		}
		
		for _, modelName := range modelNames {
			agentName := allModels[modelName]
			response.WriteString(fmt.Sprintf("• `%s` - 来自 %s\n", modelName, agentName))
		}
		
		response.WriteString("\n📝 *使用方式:*\n")
		response.WriteString("`/model <模型名称>`\n")
		response.WriteString("例如: `/model gpt-4`\n")
		response.WriteString("`/model` - 查看当前模型和可用列表\n")
	}
	
	return response.String()
}

// switchModel 切换用户模型
func (m *ModelCommand) switchModel(ctx context.Context, c telebot.Context, user *domain.User, modelName string) string {
	// 检查模型是否存在
	agentConfigs, err := getAgentConfigs()
	if err != nil {
		return fmt.Sprintf("❌ 获取模型配置失败: %v", err)
	}
	
	// 验证模型是否可用
	modelExists := false
	for _, agent := range agentConfigs {
		for _, model := range agent.Models {
			if model == modelName {
				modelExists = true
				break
			}
		}
		if modelExists {
			break
		}
	}
	
	if !modelExists {
		return fmt.Sprintf("❌ 模型 '%s' 不存在\n\n请使用 `/model` 查看可用模型列表", modelName)
	}
	
	// 更新用户模型
	if user == nil {
		return "❌ 用户信息获取失败，请先登录"
	}
	
	// 使用数据库操作更新用户模型
	err = updateUserModel(uint64(user.TelegramId), modelName)
	if err != nil {
		return fmt.Sprintf("❌ 切换模型失败: %v", err)
	}
	
	// 更新内存中的用户对象
	user.ModelName = modelName
	
	return fmt.Sprintf("✅ *模型切换成功*\n\n已切换到模型: **%s**\n\n下次对话将使用此模型", modelName)
}

// getAgentConfigs 获取所有Agent配置
func getAgentConfigs() ([]domain.AgentConfig, error) {
	// 使用工具调用获取Agent配置
	result, err := executeTool("agent_list", map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	
	// 解析JSON结果
	if strings.TrimSpace(result) == "" {
		return []domain.AgentConfig{}, nil
	}
	
	var rawAgentList []map[string]interface{}
	if err := json.Unmarshal([]byte(result), &rawAgentList); err != nil {
		// 如果不是JSON数组，尝试解析为单个对象
		var singleAgent map[string]interface{}
		if err := json.Unmarshal([]byte(result), &singleAgent); err != nil {
			return nil, fmt.Errorf("解析Agent列表失败: %v", err)
		}
		rawAgentList = []map[string]interface{}{singleAgent}
	}
	
	// 转换为domain.AgentConfig
	var agentConfigs []domain.AgentConfig
	for _, rawAgent := range rawAgentList {
		var models []string
		if modelsRaw, ok := rawAgent["models"].([]interface{}); ok {
			for _, m := range modelsRaw {
				if modelName, ok := m.(string); ok {
					models = append(models, modelName)
				}
			}
		}
		
		name := ""
		if nameRaw, ok := rawAgent["name"].(string); ok {
			name = nameRaw
		}
		
		agentConfigs = append(agentConfigs, domain.AgentConfig{
			Name:   name,
			Models: models,
		})
	}
	
	return agentConfigs, nil
}

// updateUserModel 更新用户模型
func updateUserModel(telegramId uint64, modelName string) error {
	// 使用数据库工具执行更新操作
	_, err := executeDBTool("UPDATE users SET model_name = ? WHERE telegram_id = ?", 
		[]interface{}{modelName, telegramId})
	
	return err
}

// ModelSetCommand 模型设置命令（快捷方式）
type ModelSetCommand struct{}

func (m *ModelSetCommand) Command() string {
	return "modelset"
}

func (m *ModelSetCommand) Description() string {
	return "快速设置模型（model命令的别名）"
}

func (m *ModelSetCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	// 直接调用ModelCommand的Handle方法
	modelCmd := &ModelCommand{}
	if len(args) == 0 {
		return "❌ 请指定要切换的模型名称\n📝 用法: `/modelset <模型名称>`\n示例: `/modelset gpt-4`\n\n使用 `/model` 查看可用模型列表"
	}
	return modelCmd.switchModel(ctx, c, user, args[0])
}