package file

import (
	"context"
	"encoding/json"
	"orange-agent/common"
	"orange-agent/utils/file"
)

type FileWrite struct {
	common.BaseTool
}

func (f *FileWrite) Name() string {
	return "file_write"
}

func (f *FileWrite) Description() string {
	return "Write a file"
}

func (f *FileWrite) Call(ctx context.Context, input string) (string, error) {
	jsonParams := struct {
		FilePath string `json:"file_path"`
		Content  string `json:"content"`
	}{}
	err := json.Unmarshal([]byte(input), &jsonParams)
	err = file.WriteFile(jsonParams.FilePath, jsonParams.Content)
	if err != nil {
		return "", err
	}
	return "", err
}
func (f *FileWrite) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "The path of the file to read,注意起点为" + "./",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The content to write",
			},
		},
		"required": []string{"file_path", "content"},
	}
}
