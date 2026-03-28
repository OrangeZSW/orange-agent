package file

import (
	"orange-agent/common"
	"os"
)

var FileDeleteTool = common.BaseTool{
	Name:        "file_delete",
	Description: "删除指定的文件",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "要删除的文件路径",
			},
		},
		"required": []interface{}{"file_path"},
	},
}

func DeleteFile(filePath string) (string, error) {
	err := os.Remove(filePath)
	if err != nil {
		return "", err
	}
	return "文件已成功删除：" + filePath, nil
}
