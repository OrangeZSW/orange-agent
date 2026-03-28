package agent

import (
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"orange-agent/config/config"
	"orange-agent/domain"
)

var AgentListTool = common.BaseTool{
	Name:        "agent_list",
	Description: "列出所有已配置的Agent",
	Parameters:  map[string]string{},
	Required:    []string{},
	Handler:     handleAgentList,
}

func handleAgentList(params map[string]interface{}) (string, error) {
	var agents []domain.AgentConfig
	if err := config.DB.Find(&agents).Error; err != nil {
		return "", fmt.Errorf("查询Agent列表失败: %v", err)
	}

	result := map[string]interface{}{
		"status": "success",
		"count":  len(agents),
		"data":   agents,
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}
