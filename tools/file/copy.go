package file

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"orange-agent/common"
	"os"
)

type FileCopyTools struct {
	common.BaseTool
}

func (f *FileCopyTools) Name() string {
	return "file_copy"
}

func (f *FileCopyTools) Description() string {
	return "复制文件或目录"
}

func (f *FileCopyTools) Call(ctx context.Context, input string) (string, error) {
	var params struct {
		SourcePath string `json:"source_path"`
		DestPath   string `json:"dest_path"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	if params.SourcePath == "" || params.DestPath == "" {
		return "", fmt.Errorf("source_path and dest_path are required")
	}

	sourceFile, err := os.Open(params.SourcePath)
	if err != nil {
		return "", err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(params.DestPath)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return "", err
	}

	return "文件已成功复制：" + params.SourcePath + " -> " + params.DestPath, nil
}

func (f *FileCopyTools) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"source_path": map[string]interface{}{
				"type":        "string",
				"description": "源文件路径",
			},
			"dest_path": map[string]interface{}{
				"type":        "string",
				"description": "目标文件路径",
			},
		},
		"required": []string{"source_path", "dest_path"},
	}
}
