package system

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/common"
	"os/exec"
)

var TestRunTool = common.BaseTool{
	Name:        "test_run",
	Description: "运行测试用例",
	Parameters: map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"package": map[string]interface{}{
				"type":        "string",
				"description": "要测试的包路径（可选，默认为当前目录）",
			},
			"verbose": map[string]interface{}{
				"type":        "boolean",
				"description": "是否显示详细输出",
			},
		},
	},
	Call: handlerTestRun,
}

func handlerTestRun(ctx context.Context, input string) (string, error) {
	// 解析JSON参数
	var params struct {
		Package string `json:"package"`
		Verbose bool   `json:"verbose"`
	}

	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("failed to parse arguments: %v", err)
	}

	args := []string{"test"}
	if params.Verbose {
		args = append(args, "-v")
	}
	if params.Package != "" {
		args = append(args, params.Package)
	} else {
		args = append(args, "./...")
	}

	cmd := exec.Command("go", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}
