package system

import (
	"context"
	"orange-agent/common"
	"os/exec"
)

var BuildTool = common.BaseTool{
	Name:        "build_tools",
	Description: "执行构建脚本，编译项目",
	Parameters: map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
		"required":   []string{},
	},
	Call: handlerBuild,
}

func handlerBuild(ctx context.Context, input string) (string, error) {
	cmd := exec.Command("bash", "-c", "chmod +x ./build.sh && ./build.sh")
	output, err := cmd.CombinedOutput()

	result := "Build output:\n" + string(output)
	if err != nil {
		result += "\n\nBuild failed with error: " + err.Error()
		return result, err
	}
	result += "\n\nBuild completed successfully!"
	return result, nil
}
