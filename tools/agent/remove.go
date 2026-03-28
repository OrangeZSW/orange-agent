package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"orange-agent/domain"
	"orange-agent/mysql"
	"orange-agent/utils"
)

var AgentRemoveTool = common.BaseTool{
	Name:        "agent_remove",
	Description: "删除指定Agent配置",
	Parameters: map[string]interface{}{
		"name": map[string]interface{}{
			"type":        "string",
			"description": "Agent名称",
		},
		"required": []string{"name"},
	},
	Call: handleAgentRemove,
}

func handleAgentRemove(ctx context.Context, input string) (string, error) {
	params, err := utils.StrToMap(input)
	if err != nil {
		return "", err
	}
	name := params["name"].(string)

	var agent domain.AgentConfig
	if err := mysql.GetDB().Where("name = ?", name).First(&agent).Error; err != nil {
		return "", fmt.Errorf("Agent %s 不存在", name)
	}

	if err := mysql.GetDB().Delete(&agent).Error; err != nil {
		return "", fmt.Errorf("删除Agent配置失败: %v", err)
	}

	result := map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Agent %s 删除成功", name),
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}
