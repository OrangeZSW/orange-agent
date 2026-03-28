package file

import (
	"context"
	"orange-agent/common"
	"orange-agent/utils/file"
	"strings"
)

type FileList struct {
	common.BaseTool
}

func (f *FileList) Name() string {
	return "file_list"
}

func (f *FileList) Description() string {
	return "list all files in the current directory"
}
func (f *FileList) Call(ctx context.Context, input string) (string, error) {
	fileList, err := file.GetFileTree(".")
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	var traverse func(nodes []*common.FileNode)

	traverse = func(nodes []*common.FileNode) {
		for _, node := range nodes {
			builder.WriteString(node.Path)
			builder.WriteString("\n")

			// 如果有子节点，递归遍历
			if len(node.Children) > 0 {
				traverse(node.Children)
			}
		}
	}

	traverse(fileList)
	return builder.String(), nil
}

func (f *FileList) Parameters() interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	}
}
