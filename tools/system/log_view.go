package system

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"os/exec"
	"strconv"
)

type LogViewTools struct {
	common.BaseTool
}

func (l *LogViewTools) Name() string {
	return "log_view"
}

func (l *LogViewTools) Description() string {
	return "查看应用日志文件内容"
}

func (l *LogViewTools) Call(ctx context.Context, input string) (string, error) {
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
		return "", err
	}

	return string(output), nil
}

func (l *LogViewTools) Parameters() interface{} {
	return map[string]interface{}{
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
	}
}
