package file

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"orange-agent/utils/file"
)

var FileWriteTool = common.BaseTool{
	Name:        "file_write",
	Description: "写入文件内容",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "文件路径，注意起点为 ./",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "要写入的内容",
			},
		},
		"required": []string{"file_path", "content"},
	},
	Call: handlerFileWrite,
}

func handlerFileWrite(ctx context.Context, input string) (string, error) {
	// 解析JSON参数
	var params struct {
		FilePath string `json:"file_path"`
		Content  string `json:"content"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	if params.FilePath == "" {
		return "", fmt.Errorf("file_path is required")
	}

	// 写入文件
	err := file.WriteFile(params.FilePath, params.Content)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("文件已成功写入：%s", params.FilePath), nil
}
