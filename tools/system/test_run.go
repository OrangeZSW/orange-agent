package system

import (
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
		"required": [],
	},
}

func RunTests(packagePath string, verbose bool) (string, error) {
	args := []string{"test"}
	if verbose {
		args = append(args, "-v")
	}
	if packagePath != "" {
		args = append(args, packagePath)
	} else {
		args = append(args, "./...")
	}

	cmd := exec.Command("go", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
