package file

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"os"
)

var FileRenameTool = common.BaseTool{
	Name:        "file_rename",
	Description: "重命名文件或目录",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"old_path": map[string]interface{}{
				"type":        "string",
				"description": "原文件路径",
			},
			"new_path": map[string]interface{}{
				"type":        "string",
				"description": "新文件路径",
			},
		},
		"required": []string{"old_path", "new_path"},
	},
	Call: handlerFileRename,
}

func handlerFileRename(ctx context.Context, input string) (string, error) {
	// 解析JSON参数
	var params struct {
		OldPath string `json:"old_path"`
		NewPath string `json:"new_path"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	if params.OldPath == "" || params.NewPath == "" {
		return "", fmt.Errorf("old_path and new_path are required")
	}

	// 执行重命名操作
	err := os.Rename(params.OldPath, params.NewPath)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("文件已成功重命名：%s -> %s", params.OldPath, params.NewPath), nil
}
