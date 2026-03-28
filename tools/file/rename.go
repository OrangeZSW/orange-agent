package file

import (
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
		"required": []interface{}{"old_path", "new_path"},
	},
}

func RenameFile(oldPath, newPath string) (string, error) {
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return "", err
	}
	return "文件已成功重命名：" + oldPath + " -> " + newPath, nil
}
