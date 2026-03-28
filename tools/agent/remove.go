package agent

import (
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"orange-agent/config/config"
	"orange-agent/domain"
	"orange-agent/mysql"
)

var AgentRemoveTool = common.BaseTool{
	Name:        "agent_remove",
	Description: "删除指定Agent配置",
	Parameters: map[string]string{
		"name": "Agent名称",
	},
	Required: []string{"name"},
	Handler:  handleAgentRemove,
}

func handleAgentRemove(params map[string]interface{}) (string, error) {
	name, ok := params["name"].(string)
	if !ok || name == "" {
		return "", fmt.Errorf("name 参数不能为空")
	}

	var agent domain.AgentConfig
	if err := config.DB.Where("name = ?", name).First(&agent).Error; err != nil {
		return "", fmt.Errorf("Agent %s 不存在", name)
	}

	if err := config.DB.Delete(&agent).Error; err != nil {
		return "", fmt.Errorf("删除Agent配置失败: %v", err)
	}

	// 刷新缓存
	mysql.LoadAgentCache()

	result := map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Agent %s 删除成功", name),
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}
