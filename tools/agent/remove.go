package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	factory "orange-agent/repository/factory"
	"orange-agent/utils"
)

var AgentRemoveTool = common.BaseTool{
	Name:        "agent_remove",
	Description: "删除指定Agent配置",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Agent名称",
			},
		},
		"required": []string{"name"},
	},
	Call: handleAgentRemove,
}

func handleAgentRemove(ctx context.Context, input string) (string, error) {
	repo := factory.NewFactory()
	params, err := utils.StrToMap(input)
	if err != nil {
		return "", err
	}
	name := params["name"].(string)

	agent, err := repo.AgentConfigRepo.GetAgentConfigByName(name)
	if err != nil {
		return "", fmt.Errorf("Agent %s 不存在", name)
	}

	if err := repo.AgentConfigRepo.DeleteAgentConfig(agent); err != nil {
		return "", fmt.Errorf("删除Agent配置失败: %v", err)
	}

	result := map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Agent %s 删除成功", name),
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}
