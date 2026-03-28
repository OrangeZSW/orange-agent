package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"orange-agent/domain"
	"orange-agent/mysql"
)

var AgentListTool = common.BaseTool{
	Name:        "agent_list",
	Description: "列出所有已配置的Agent",
	Parameters:  map[string]interface{}{},
	Call:        handleAgentList,
}

func handleAgentList(ctx context.Context, input string) (string, error) {
	var agents []domain.AgentConfig
	if err := mysql.GetDB().Find(&agents).Error; err != nil {
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
