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
	Description: "当文件更新，新增，删除，时执行增量更新，当索引文件不存在时，执行全量更新",
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
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		params.ProjectRoot = "./"
	}

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
		err = rag.InitializeIndex(ctx, params.ProjectRoot)
		modeName = "全量重建"
	case "incremental":
		err = rag.InitializeIndexIncremental(ctx, params.ProjectRoot)
		modeName = "增量更新"
	default:
		return "", fmt.Errorf("不支持的模式: %s，可选值: full, incremental", params.Mode)
	}

	if err != nil {
		return "", err
	}

	retriever := rag.GetRetriever()
	size, _ := retriever.GetIndexSize(ctx)
	return fmt.Sprintf("代码索引%s完成，共 %d 个代码块。现在可以使用 code_search 搜索代码。", modeName, size), nil
}
