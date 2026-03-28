package system

import (
	"orange-agent/common"
	"os"
)

var EnvManageTool = common.BaseTool{
	Name:        "env_manage",
	Description: "管理环境变量（获取或设置）",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "操作类型：get（获取）、set（设置）、list（列出所有）",
				"enum":        []interface{}{"get", "set", "list"},
			},
			"key": map[string]interface{}{
				"type":        "string",
				"description": "环境变量名（当action为get或set时需要）",
			},
			"value": map[string]interface{}{
				"type":        "string",
				"description": "环境变量值（当action为set时需要）",
			},
		},
		"required": []interface{}{"action"},
	},
}

func ManageEnv(action, key, value string) (string, error) {
	switch action {
	case "list":
		envs := os.Environ()
		result := "当前环境变量:\n"
		for _, env := range envs {
			result += env + "\n"
		}
		return result, nil

	case "get":
		if key == "" {
			return "", nil
		}
		val := os.Getenv(key)
		if val == "" {
			return "环境变量 " + key + " 未设置", nil
		}
		return key + "=" + val, nil

	case "set":
		if key == "" || value == "" {
			return "", nil
		}
		err := os.Setenv(key, value)
		if err != nil {
			return "", err
		}
		return "已设置环境变量：" + key + "=" + value, nil

	default:
		return "无效的操作类型", nil
	}
}
