package system

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"os"
)

var EnvManageTool = common.BaseTool{
	Name:        "env_manage",
	Description: "管理环境变量（获取或设置）",
	Parameters: map[string]interface{}{
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
		"required": []string{"action"},
	},
	Call: handlerEnvManage,
}

func handlerEnvManage(ctx context.Context, input string) (string, error) {
	// 解析JSON参数
	var params struct {
		Action string `json:"action"`
		Key    string `json:"key"`
		Value  string `json:"value"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	if params.Action == "" {
		return "", fmt.Errorf("action is required")
	}

	switch params.Action {
	case "list":
		envs := os.Environ()
		result := "当前环境变量:\n"
		for _, env := range envs {
			result += env + "\n"
		}
		return result, nil

	case "get":
		if params.Key == "" {
			return "", fmt.Errorf("key is required for get action")
		}
		val := os.Getenv(params.Key)
		if val == "" {
			return fmt.Sprintf("环境变量 %s 未设置", params.Key), nil
		}
		return fmt.Sprintf("%s=%s", params.Key, val), nil

	case "set":
		if params.Key == "" || params.Value == "" {
			return "", fmt.Errorf("key and value are required for set action")
		}
		err := os.Setenv(params.Key, params.Value)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("已设置环境变量：%s=%s", params.Key, params.Value), nil

	default:
		return "", fmt.Errorf("invalid action: %s", params.Action)
	}
}
