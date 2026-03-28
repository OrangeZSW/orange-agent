package system

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"os"
)

type EnvManageTools struct {
	common.BaseTool
}

func (e *EnvManageTools) Name() string {
	return "env_manage"
}

func (e *EnvManageTools) Description() string {
	return "管理环境变量（获取或设置）"
}

func (e *EnvManageTools) Call(ctx context.Context, input string) (string, error) {
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
			return "环境变量 " + params.Key + " 未设置", nil
		}
		return params.Key + "=" + val, nil

	case "set":
		if params.Key == "" || params.Value == "" {
			return "", fmt.Errorf("key and value are required for set action")
		}
		err := os.Setenv(params.Key, params.Value)
		if err != nil {
			return "", err
		}
		return "已设置环境变量：" + params.Key + "=" + params.Value, nil

	default:
		return "", fmt.Errorf("invalid action: %s", params.Action)
	}
}

func (e *EnvManageTools) Parameters() interface{} {
	return map[string]interface{}{
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
		"required": []string{"action"},
	}
}
