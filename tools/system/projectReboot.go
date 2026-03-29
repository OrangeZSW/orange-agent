package system

import (
	"context"
	"orange-agent/common"
	"os/exec"
	"time"
)

var ProjectRebootTool = common.BaseTool{
	Name:        "project_reboot",
	Description: "重启项目 - 重新编译并运行，需要确认才能执行",
	Parameters:  map[string]interface{}{},
	Call:        handlerProjectReboot,
}

func handlerProjectReboot(ctx context.Context, input string) (string, error) {

	// 使用异步重启，让当前进程先退出
	go func() {
		time.Sleep(1 * time.Second) // 等待当前响应返回
		cmd := exec.Command("./start.sh")
		cmd.Start() // 使用 Start 而不是 Run，不等待完成
	}()

	return "正在重启服务...", nil
}
