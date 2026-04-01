package command

import (
	"context"
	"fmt"
	"orange-agent/agent/tools/system"
	"orange-agent/domain"
	"strings"

	"gopkg.in/telebot.v3"
)

// BuildCommand 构建项目命令
type BuildCommand struct{}

func (b *BuildCommand) Command() string {
	return "build"
}

func (b *BuildCommand) Description() string {
	return "构建项目"
}

func (b *BuildCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	// 执行构建操作
	result, err := system.BuildTools()
	if err != nil {
		return fmt.Sprintf("❌ 构建失败: %v", err)
	}
	
	var response strings.Builder
	response.WriteString("🔧 *项目构建*\n\n")
	
	// 检查构建结果
	if strings.Contains(result, "successfully") || strings.Contains(result, "成功") || strings.Contains(result, "BUILD SUCCESSFUL") {
		response.WriteString("✅ *构建成功*\n\n")
	} else if strings.Contains(result, "error") || strings.Contains(result, "失败") || strings.Contains(result, "BUILD FAILED") {
		response.WriteString("❌ *构建失败*\n\n")
	}
	
	response.WriteString("```\n")
	response.WriteString(result)
	response.WriteString("\n```")
	
	return response.String()
}

// TestCommand 运行测试命令
type TestCommand struct{}

func (t *TestCommand) Command() string {
	return "test"
}

func (t *TestCommand) Description() string {
	return "运行项目测试"
}

func (t *TestCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	var packagePath string
	if len(args) > 0 {
		packagePath = args[0]
	}
	
	// 执行测试操作
	result, err := system.TestRun(packagePath, false)
	if err != nil {
		return fmt.Sprintf("❌ 测试运行失败: %v", err)
	}
	
	var response strings.Builder
	response.WriteString("🧪 *测试结果*\n\n")
	
	// 分析测试结果
	if strings.Contains(result, "PASS") || strings.Contains(result, "ok") {
		response.WriteString("✅ *测试通过*\n\n")
	} else if strings.Contains(result, "FAIL") || strings.Contains(result, "错误") {
		response.WriteString("❌ *测试失败*\n\n")
	}
	
	response.WriteString("```\n")
	response.WriteString(result)
	response.WriteString("\n```")
	
	return response.String()
}

// RebootCommand 重启项目命令
type RebootCommand struct{}

func (r *RebootCommand) Command() string {
	return "reboot"
}

func (r *RebootCommand) Description() string {
	return "重启项目"
}

func (r *RebootCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	// 执行项目重启操作
	result, err := system.ProjectReboot()
	if err != nil {
		return fmt.Sprintf("❌ 重启失败: %v", err)
	}
	
	return fmt.Sprintf("🔄 *项目重启*\n\n%s", result)
}

// DependencyCheckCommand 依赖检查命令
type DependencyCheckCommand struct{}

func (d *DependencyCheckCommand) Command() string {
	return "deps"
}

func (d *DependencyCheckCommand) Description() string {
	return "检查项目依赖"
}

func (d *DependencyCheckCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	// 执行依赖检查操作
	result, err := system.DependencyCheck(false, false)
	if err != nil {
		return fmt.Sprintf("❌ 依赖检查失败: %v", err)
	}
	
	var response strings.Builder
	response.WriteString("📦 *依赖检查结果*\n\n")
	response.WriteString("```\n")
	response.WriteString(result)
	response.WriteString("\n```")
	
	return response.String()
}

// LogViewCommand 查看日志命令
type LogViewCommand struct{}

func (l *LogViewCommand) Command() string {
	return "logs"
}

func (l *LogViewCommand) Description() string {
	return "查看应用日志"
}

func (l *LogViewCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	// 默认查看最新50行日志
	lines := 50
	if len(args) > 0 {
		// 可以支持指定行数，这里简化处理
		// 实际可以解析参数如: /logs 100
	}
	
	// 执行日志查看操作
	result, err := system.LogView("./log/orange-agent.log", lines)
	if err != nil {
		return fmt.Sprintf("❌ 查看日志失败: %v", err)
	}
	
	var response strings.Builder
	response.WriteString("📝 *应用日志 (最新50行)*\n\n")
	response.WriteString("```\n")
	response.WriteString(result)
	response.WriteString("\n```")
	
	return response.String()
}