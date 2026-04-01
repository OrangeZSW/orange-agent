package system

import (
	"context"
	"orange-agent/common"
	"os"
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
	// 首先检查 build.sh 是否存在
	if _, err := os.Stat("./build.sh"); os.IsNotExist(err) {
		return "错误: build.sh 文件不存在，尝试直接使用 go build", nil
	}

	// 确保脚本有执行权限
	if err := os.Chmod("./build.sh", 0755); err != nil {
		return "警告: 无法修改 build.sh 权限: " + err.Error(), nil
	}

	cmd := exec.CommandContext(ctx, "bash", "./build.sh")
	output, err := cmd.CombinedOutput()

	result := "=== 构建输出 ===\n" + string(output)
	if err != nil {
		result += "\n\n❌ 构建失败: " + err.Error()
		return result, nil
	}
	result += "\n\n✅ 构建成功完成!"
	return result, nil
}
