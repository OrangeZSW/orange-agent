package file

import (
	"orange-agent/common"
	"os"
)

var FileCopyTool = common.BaseTool{
	Name:        "file_copy",
	Description: "复制文件或目录",
	Parameters: map[string]interface{}{
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
		"required": []interface{}{"source_path", "dest_path"},
	},
}

func CopyFile(sourcePath, destPath string) (string, error) {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return "", err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	_, err = copyContent(sourceFile, destFile)
	if err != nil {
		return "", err
	}

	return "文件已成功复制：" + sourcePath + " -> " + destPath, nil
}

func copyContent(src *os.File, dst *os.File) (int64, error) {
	return src.WriteTo(dst)
}
