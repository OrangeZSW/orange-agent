package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"orange-agent/repository/resource"
	"strings"
)

var AgentUpdateTool = common.BaseTool{
	Name:        "agent_update",
	Description: "更新Agent配置信息",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Agent名称",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "Agent类型 (可选)",
			},
			"model": map[string]interface{}{
				"type":        "string",
				"description": "默认模型名称，多个模型用逗号分隔 (可选)",
			},
			"endpoint": map[string]interface{}{
				"type":        "string",
				"description": "API端点URL (可选)",
			},
			"api_key": map[string]interface{}{
				"type":        "string",
				"description": "API密钥 (可选)",
			},
		},
		"required": []string{"name"},
	},
	Call: handlerAgentUpdate,
}

func handlerAgentUpdate(ctx context.Context, input string) (string, error) {
	repo := resource.GetRepositories()
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

	if params.Name == "" {
		return "", fmt.Errorf("name is required")
	}

	agent, err := repo.AgentConfig.GetAgentConfigByName(params.Name)
	if err != nil {
		return "", fmt.Errorf("Agent %s 不存在", params.Name)
	}

	// 更新可修改字段
	if params.Endpoint != "" {
		agent.BaseUrl = params.Endpoint
	}
	if params.ApiKey != "" {
		agent.Token = params.ApiKey
	}
	if params.Model != "" {
		agent.Models = strings.Split(params.Model, ",")
	}

	if err := repo.AgentConfig.UpdateAgentConfig(agent); err != nil {
		return "", fmt.Errorf("更新Agent配置失败: %v", err)
	}

	result := map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Agent %s 更新成功", params.Name),
		"data":    agent,
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}
