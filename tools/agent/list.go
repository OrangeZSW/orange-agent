package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	repo_factory "orange-agent/repository/factory"
)

var AgentListTool = common.BaseTool{
	Name:        "agent_list",
	Description: "列出所有已配置的Agent",
	Parameters:  map[string]interface{}{},
	Call:        handleAgentList,
}

func handleAgentList(ctx context.Context, input string) (string, error) {
	repo := repo_factory.NewFactory()
	agents, err := repo.AgentConfigRepo.GetAllAgentConfig()
	if err != nil {
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
