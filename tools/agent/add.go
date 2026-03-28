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

var AgentAddTool = common.BaseTool{
	Name:        "agent_add",
	Description: "添加新的Agent配置",
	Parameters: map[string]string{
		"name":     "Agent名称",
		"type":     "Agent类型 (doubao/openai/other)",
		"model":    "默认模型名称",
		"endpoint": "API端点URL",
		"api_key":  "API密钥",
	},
	Required: []string{"name", "endpoint", "api_key"},
	Handler:  handleAgentAdd,
}

func handleAgentAdd(params map[string]interface{}) (string, error) {
	name, ok := params["name"].(string)
	if !ok || name == "" {
		return "", fmt.Errorf("name 参数不能为空")
	}

	agentType, _ := params["type"].(string)
	model, _ := params["model"].(string)
	endpoint, ok := params["endpoint"].(string)
	if !ok || endpoint == "" {
		return "", fmt.Errorf("endpoint 参数不能为空")
	}

	apiKey, ok := params["api_key"].(string)
	if !ok || apiKey == "" {
		return "", fmt.Errorf("api_key 参数不能为空")
	}

	// 检查是否已存在
	var existing domain.AgentConfig
	if err := config.DB.Where("name = ?", name).First(&existing).Error; err == nil {
		return "", fmt.Errorf("Agent %s 已存在", name)
	}

	var models []string
	if model != "" {
		models = strings.Split(model, ",")
	}

	agentConfig := domain.AgentConfig{
		Name:    name,
		Token:   apiKey,
		BaseUrl: endpoint,
		Models:  models,
	}

	if err := config.DB.Create(&agentConfig).Error; err != nil {
		return "", fmt.Errorf("创建Agent配置失败: %v", err)
	}

	// 刷新缓存
	mysql.LoadAgentCache()

	result := map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Agent %s 添加成功", name),
		"data":    agentConfig,
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}
