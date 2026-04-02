package system

import (
	"context"
	"encoding/json"
	"orange-agent/agent/rag"
	"orange-agent/common"
)

var CodeSearchTool = common.BaseTool{
	Name:        "code_search",
	Description: "搜索项目代码库，根据问题检索相关代码片段。当你需要理解项目代码、查找特定功能实现、分析代码逻辑时使用此工具。返回相关的代码片段及其位置。",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "搜索查询，描述你要查找的代码功能或关键词",
			},
			"top_k": map[string]interface{}{
				"type":        "integer",
				"description": "返回结果数量，默认5个",
			},
		},
		"required": []string{"query"},
	},
	Call: handlerCodeSearch,
}

func handlerCodeSearch(ctx context.Context, input string) (string, error) {
	var params struct {
		Query string `json:"query"`
		TopK  int    `json:"top_k"`
	}
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", err
	}

	// 设置默认值
	if params.TopK <= 0 {
		params.TopK = 5
	}

	retriever := rag.GetRetriever()
	chunks, err := retriever.Retrieve(ctx, params.Query, params.TopK)
	if err != nil {
		return "", err
	}

	if len(chunks) == 0 {
		return "未找到相关代码。可能需要先初始化代码索引，请使用 code_index_init 工具。", nil
	}

	return retriever.BuildContext(chunks), nil
}
