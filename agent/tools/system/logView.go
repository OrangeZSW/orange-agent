package system

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"os/exec"
	"strconv"
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
		"required": []string{"log_file"},
	},
	Call: handlerLogView,
}

func handlerLogView(ctx context.Context, input string) (string, error) {
	// 解析JSON参数
	var params struct {
		LogFile string `json:"log_file"`
		Lines   int    `json:"lines"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	if params.LogFile == "" {
		return "", fmt.Errorf("log_file is required")
	}

	lines := params.Lines
	if lines <= 0 {
		lines = 50
	}

	cmd := exec.Command("tail", "-n", strconv.Itoa(lines), params.LogFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to read log file: %v\n%s", err, string(output))
	}

	return string(output), nil
}
