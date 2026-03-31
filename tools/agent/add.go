package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"orange-agent/domain"
	factory "orange-agent/repository/factory"
	"strings"
)

var AgentAddTool = common.BaseTool{
	Name:        "agent_add",
	Description: "添加新的Agent配置",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Agent名称",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Agent类型 (doubao/openai/other)",
			},
			"model": map[string]interface{}{
				"type":        "string",
				"description": "默认模型名称，多个模型用逗号分隔",
			},
			"endpoint": map[string]interface{}{
				"type":        "string",
				"description": "API端点URL",
			},
			"api_key": map[string]interface{}{
				"type":        "string",
				"description": "API密钥",
			},
		},
		"required": []string{"name", "endpoint", "api_key"},
	},
	Call: handlerAgentAdd,
}

func handlerAgentAdd(ctx context.Context, input string) (string, error) {
	factory := factory.NewFactory()
	// 解析JSON参数
	var params struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Model    string `json:"model"`
		Endpoint string `json:"endpoint"`
		ApiKey   string `json:"api_key"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	// 参数验证
	if params.Name == "" {
		return "", fmt.Errorf("name is required")
	}
	if params.Endpoint == "" {
		return "", fmt.Errorf("endpoint is required")
	}
	if params.ApiKey == "" {
		return "", fmt.Errorf("api_key is required")
	}

	// 检查是否已存在
	existingAgent, err := factory.AgentConfigRepo.GetAgentConfigByName(params.Name)
	if err != nil {
		return "", fmt.Errorf("failed to check existing Agent: %v", err)
	}
	if existingAgent != nil {
		return "", fmt.Errorf("Agent with name %s already exists", params.Name)
	}

	// 处理模型列表
	var models []string
	if params.Model != "" {
		models = strings.Split(params.Model, ",")
	}

	agentConfig := domain.AgentConfig{
		Name:    params.Name,
		Token:   params.ApiKey,
		BaseUrl: params.Endpoint,
		Models:  models,
	}

	if err := factory.AgentConfigRepo.CreateAgentConfig(&agentConfig); err != nil {
		return "", fmt.Errorf("创建Agent配置失败: %v", err)
	}

	result := map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Agent %s 添加成功", params.Name),
		"data":    agentConfig,
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}
