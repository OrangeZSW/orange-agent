package system

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"os"

	"gopkg.in/yaml.v3"
)

type ConfigValidatorTools struct {
	common.BaseTool
}

func (c *ConfigValidatorTools) Name() string {
	return "config_validator"
}

func (c *ConfigValidatorTools) Description() string {
	return "验证配置文件格式和语法"
}

func (c *ConfigValidatorTools) Call(ctx context.Context, input string) (string, error) {
	var params struct {
		ConfigFile string `json:"config_file"`
		Format     string `json:"format"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	if params.ConfigFile == "" {
		return "", fmt.Errorf("config_file is required")
	}

	content, err := os.ReadFile(params.ConfigFile)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	format := params.Format
	if format == "" {
		// 尝试从文件扩展名推断
		if len(params.ConfigFile) > 5 && params.ConfigFile[len(params.ConfigFile)-5:] == ".yaml" {
			format = "yaml"
		} else if len(params.ConfigFile) > 5 && params.ConfigFile[len(params.ConfigFile)-5:] == ".yml" {
			format = "yaml"
		} else if len(params.ConfigFile) > 5 && params.ConfigFile[len(params.ConfigFile)-5:] == ".json" {
			format = "json"
		} else {
			return "", fmt.Errorf("无法推断配置文件格式，请指定 format 参数")
		}
	}

	var result string
	switch format {
	case "yaml":
		var data interface{}
		if err := yaml.Unmarshal(content, &data); err != nil {
			return "", fmt.Errorf("YAML 格式错误: %v", err)
		}
		result = "YAML 配置文件格式正确"
	case "json":
		var data interface{}
		if err := json.Unmarshal(content, &data); err != nil {
			return "", fmt.Errorf("JSON 格式错误: %v", err)
		}
		result = "JSON 配置文件格式正确"
	case "toml":
		// 简单检查，实际应该使用 toml 库
		result = "TOML 格式验证（简化检查）- 文件存在且可读"
	default:
		return "", fmt.Errorf("不支持的格式: %s", format)
	}

	return result + "\n文件: " + params.ConfigFile, nil
}

func (c *ConfigValidatorTools) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"config_file": map[string]interface{}{
				"type":        "string",
				"description": "配置文件路径（例如：./config.yaml）",
			},
			"format": map[string]interface{}{
				"type":        "string",
				"description": "配置格式：yaml、json、toml等",
				"enum":        []interface{}{"yaml", "json", "toml"},
			},
		},
		"required": []string{"config_file"},
	}
}
