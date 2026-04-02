package system

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/agent/rag"
	"orange-agent/common"
)

var CodeIndexInitTool = common.BaseTool{
	Name:        "code_index_init",
	Description: "初始化或刷新代码索引。增量模式只处理变化的文件，速度更快。",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_root": map[string]interface{}{
				"type":        "string",
				"description": "项目根目录路径，默认为当前目录 './'",
			},
			"mode": map[string]interface{}{
				"type":        "string",
				"description": "索引模式: 'full' 全量重建, 'incremental' 增量更新（默认）",
				"enum":        []string{"full", "incremental"},
			},
		},
		"required": []string{},
	},
	Call: handlerCodeIndexInit,
}

func handlerCodeIndexInit(ctx context.Context, input string) (string, error) {
	var params struct {
		ProjectRoot string `json:"project_root"`
		Mode        string `json:"mode"`
	}
	json.Unmarshal([]byte(input), &params)

	if params.ProjectRoot == "" {
		params.ProjectRoot = "./"
	}
	if params.Mode == "" {
		params.Mode = "incremental"
	}

	var err error
	var modeName string

	switch params.Mode {
	case "full":
		err = rag.IndexFull(ctx, params.ProjectRoot)
		modeName = "全量重建"
	case "incremental":
		err = rag.IndexIncremental(ctx, params.ProjectRoot)
		modeName = "增量更新"
	default:
		return "", fmt.Errorf("不支持的模式: %s", params.Mode)
	}

	if err != nil {
		return "", err
	}

	size, _ := rag.GetSize(ctx)
	return fmt.Sprintf("代码索引%s完成，共 %d 个代码块", modeName, size), nil
}
