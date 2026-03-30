package file

import (
	"context"
	"orange-agent/common"
	"orange-agent/utils"
	"orange-agent/utils/file"
)

var RandomReadFile = common.BaseTool{

	Name:        "randomReadFile",
	Description: "随机读取文件内容，参数为文件路径和偏移量，返回文件内容",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"file_path": map[string]interface{}{
				"type":        "string",
				"description": "文件路径",
			},
			"offset": map[string]interface{}{
				"type":        "string",
				"description": "偏移量",
			},
			"length": map[string]interface{}{
				"type":        "string",
				"description": "读取长度",
			},
		},
		"required": []string{"file_path", "offset", "length"},
	},
	Call: handleRandomReadFile,
}

func handleRandomReadFile(ctx context.Context, input string) (string, error) {
	params, _ := utils.StrToMap(input)

	byte, err := file.ReadRandomAccess(params["file_path"].(string), utils.StrToInt64(params["offset"].(string)), utils.StrToInt64(params["length"].(string)))
	if err != nil {
		return "", err
	}
	return string(byte), nil
}
