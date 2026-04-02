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
	Description: "初始化或刷新代码索引。当需要搜索代码但索引未初始化时使用。会对项目代码进行向量化索引，以便后续使用 code_search 搜索。",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"project_root": map[string]interface{}{
				"type":        "string",
				"description": "项目根目录路径，默认为当前目录 './'",
			},
		},
		"required": []string{},
	},
	Call: handlerCodeIndexInit,
}

func handlerCodeIndexInit(ctx context.Context, input string) (string, error) {
	var params struct {
		ProjectRoot string `json:"project_root"`
	}
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		// 如果解析失败，使用默认值
		params.ProjectRoot = "./"
	}

	// 设置默认值
	if params.ProjectRoot == "" {
		params.ProjectRoot = "./"
	}

	if err := rag.InitializeIndex(ctx, params.ProjectRoot); err != nil {
		return "", err
	}

	retriever := rag.GetRetriever()
	size, _ := retriever.GetIndexSize(ctx)
	return "代码索引初始化完成，共索引 " + fmt.Sprintf("%d", size) + " 个代码块。现在可以使用 code_search 工具搜索代码了。", nil
}
