package system

import (
	"context"
	"orange-agent/common"
	"os/exec"
	"time"
)

type ProjectReboot struct {
	common.BaseTool
}

func (p *ProjectReboot) Name() string {
	return "project_reboot"
}

func (p *ProjectReboot) Description() string {
	return "重启项目 - 重新编译并运行，需要确认才能执行"
}

func (p *ProjectReboot) Call(ctx context.Context, input string) (string, error) {
	// 使用异步重启，让当前进程先退出
	go func() {
		time.Sleep(1 * time.Second) // 等待当前响应返回
		cmd := exec.Command("./start.sh")
		cmd.Start() // 使用 Start 而不是 Run，不等待完成
	}()

	return "正在重启服务...", nil
}

func (p *ProjectReboot) Parameters() interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"confirm": map[string]interface{}{
				"type":        "boolean",
				"description": "确认是否要重启项目",
			},
		},
		"required": []string{"confirm"},
	}
}
