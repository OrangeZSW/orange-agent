package system

import (
	"orange-agent/common"
	"os/exec"
)

var LogViewTool = common.BaseTool{
	Name:        "log_view",
	Description: "查看应用日志文件内容",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"log_file": map[string]interface{}{
				"type":        "string",
				"description": "日志文件路径（例如：./log/orange-agent.log）",
			},
			"lines": map[string]interface{}{
				"type":        "integer",
				"description": "要查看的行数（可选，默认为50行）",
			},
		},
		"required": []interface{}{"log_file"},
	},
}

func ViewLog(logFile string, lines int) (string, error) {
	if lines <= 0 {
		lines = 50
	}

	cmd := exec.Command("tail", "-n", string(rune(lines)), logFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
