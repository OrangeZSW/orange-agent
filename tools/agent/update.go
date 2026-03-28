package agent

import (
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"orange-agent/config/config"
	"orange-agent/domain"
	"orange-agent/mysql"
	"strings"
)

var AgentUpdateTool = common.BaseTool{
	Name:        "agent_update",
	Description: "更新Agent配置信息",
	Parameters: map[string]string{
		"name":     "Agent名称",
		"type":     "Agent类型 (可选)",
		"model":    "默认模型名称 (可选)",
		"endpoint": "API端点URL (可选)",
		"api_key":  "API密钥 (可选)",
	},
	Required: []string{"name"},
	Handler:  handleAgentUpdate,
}

func handleAgentUpdate(params map[string]interface{}) (string, error) {
	name, ok := params["name"].(string)
	if !ok || name == "" {
		return "", fmt.Errorf("name 参数不能为空")
	}

	var agent domain.AgentConfig
	if err := config.DB.Where("name = ?", name).First(&agent).Error; err != nil {
		return "", fmt.Errorf("Agent %s 不存在", name)
	}

	// 更新可修改字段
	if endpoint, ok := params["endpoint"].(string); ok && endpoint != "" {
		agent.BaseUrl = endpoint
	}
	if apiKey, ok := params["api_key"].(string); ok && apiKey != "" {
		agent.Token = apiKey
	}
	if model, ok := params["model"].(string); ok && model != "" {
		agent.Models = strings.Split(model, ",")
	}

	if err := config.DB.Save(&agent).Error; err != nil {
		return "", fmt.Errorf("更新Agent配置失败: %v", err)
	}

	// 刷新缓存
	mysql.LoadAgentCache()

	result := map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Agent %s 更新成功", name),
		"data":    agent,
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}
