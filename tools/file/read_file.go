package file

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"orange-agent/utils/file"
)

var FileReadTool = common.BaseTool{
	Name:        "file_read",
	Description: "用于读取文件内容",
	Parameters: map[string]interface{}{
		"file_path": map[string]interface{}{
			"type":        "string",
			"description": "文件路径",
		},
		"required": []string{"file_path"},
	},
	Call: handlerFileRead,
}

func handlerFileRead(ctx context.Context, input string) (string, error) {
	// 解析JSON参数
	var params struct {
		FilePath string `json:"file_path"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	if params.FilePath == "" {
		return "", fmt.Errorf("file_path is required")
	}

	// 使用解析出的文件路径
	content, err := file.ReadFile(params.FilePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
