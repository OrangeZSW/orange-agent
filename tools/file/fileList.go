package file

import (
	"context"
	"orange-agent/common"
	"orange-agent/utils/file"
	"strings"
)

var FileListTool = common.BaseTool{
	Name:        "file_list",
	Description: "list all files in the current directory",
	Call:        handlerFileList,
	Parameters: map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	},
}

func handlerFileList(ctx context.Context, input string) (string, error) {
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
