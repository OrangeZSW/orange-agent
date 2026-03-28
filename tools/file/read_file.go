package file

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"orange-agent/utils/file"
)

type FileRead struct {
	common.BaseTool
}

func (f *FileRead) Name() string {
	return "file_read"
}

// 小文件处理
func (f *FileRead) Description() string {
	return "Read a file and return the content,"
}

func (f *FileRead) Call(ctx context.Context, input string) (string, error) {
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
func (f *FileRead) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "The path of the file to read,注意起点为" + "./",
			},
		},
		"required": []string{"file_path"},
	}
}
