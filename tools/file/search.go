package file

import (
	"orange-agent/common"
	"os"
	"path/filepath"
	"strings"
)

var FileSearchTool = common.BaseTool{
	Name:        "file_search",
	Description: "在项目中搜索包含特定内容的文件",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"pattern": map[string]interface{}{
				"type":        "string",
				"description": "要搜索的内容或文件名模式",
			},
			"directory": map[string]interface{}{
				"type":        "string",
				"description": "搜索的目录（可选，默认为当前目录）",
			},
		},
		"required": []interface{}{"pattern"},
	},
}

func SearchFiles(pattern, directory string) (string, error) {
	if directory == "" {
		directory = "."
	}

	var results []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// 检查文件名是否匹配
		if strings.Contains(info.Name(), pattern) {
			results = append(results, path)
			return nil
		}

		// 如果是文本文件，检查内容是否包含模式
		if isTextFile(info.Name()) {
			content, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			if strings.Contains(string(content), pattern) {
				results = append(results, path)
			}
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if len(results) == 0 {
		return "未找到匹配的文件", nil
	}

	result := "找到以下文件:\n"
	for _, r := range results {
		result += "- " + r + "\n"
	}

	return result, nil
}

func isTextFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	textExtensions := []string{".go", ".py", ".js", ".ts", ".java", ".c", ".cpp", ".h", ".hpp", ".txt", ".md", ".yaml", ".yml", ".json", ".xml", ".html", ".css"}
	for _, extName := range textExtensions {
		if ext == extName {
			return true
		}
	}
	return false
}
