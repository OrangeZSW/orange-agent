package file

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"os"
)

var FileDeleteTool = common.BaseTool{
	Name:        "file_delete",
	Description: "删除指定的文件",
	Call:        handlerFileDelete,
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "要删除的文件路径",
			},
			"force": map[string]interface{}{
				"type":        "boolean",
				"description": "是否强制删除，默认为false",
			},
		},
		"required": []string{"file_path"},
	},
}

func handlerFileDelete(ctx context.Context, input string) (string, error) {
	var params struct {
		FilePath string `json:"file_path"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	if params.FilePath == "" {
		return "", fmt.Errorf("file_path is required")
	}

	err := os.Remove(params.FilePath)
	if err != nil {
		return "", err
	}
	return "文件已成功删除：" + params.FilePath, nil
}
