package file

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"os"
)

type FileRenameTools struct {
	common.BaseTool
}

func (f *FileRenameTools) Name() string {
	return "file_rename"
}

func (f *FileRenameTools) Description() string {
	return "重命名文件或目录"
}

func (f *FileRenameTools) Call(ctx context.Context, input string) (string, error) {
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

	err := os.Rename(params.OldPath, params.NewPath)
	if err != nil {
		return "", err
	}
	return "文件已成功重命名：" + params.OldPath + " -> " + params.NewPath, nil
}

func (f *FileRenameTools) Parameters() interface{} {
	return map[string]interface{}{
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
	}
}
