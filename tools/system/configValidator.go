package system

import (
	"context"
	"fmt"
	"orange-agent/common"
	"os/exec"
)

var BuildTool = common.BaseTool{
	Name:        "build_tools",
	Description: "执行构建脚本，编译项目",
	Parameters:  map[string]interface{}{},
	Call:        handlerBuild,
}

func handlerBuild(ctx context.Context, input string) (string, error) {
	cmd := exec.Command("bash", "-c", "chmod +x ./build.sh && ./build.sh")
	output, err := cmd.CombinedOutput()

	result := "Build output:\n" + string(output)
	if err != nil {
		result += fmt.Sprintf("\n\nBuild failed with error: %v", err)
		return result, err
	}
	result += "\n\nBuild completed successfully!"
	return result, nil
}
