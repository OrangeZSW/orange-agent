package command

import (
	"context"
	"fmt"
	"orange-agent/domain"
	"strings"

	"gopkg.in/telebot.v3"
)

// BuildCommand 构建命令
type BuildCommand struct{}

func (b *BuildCommand) Command() string {
	return "build"
}

func (b *BuildCommand) Description() string {
	return "构建项目"
}

func (b *BuildCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	// 执行构建操作
	result, err := executeTool("build_tools", map[string]interface{}{})
	if err != nil {
		return fmt.Sprintf("❌ 构建失败: %v", err)
	}
	
	// 添加构建状态指示
	var prefix string
	if strings.Contains(result, "✅ 构建成功") || strings.Contains(result, "构建成功完成") {
		prefix = "✅ *构建成功*\n\n"
	} else if strings.Contains(result, "❌ 构建失败") || strings.Contains(result, "错误") {
		prefix = "❌ *构建失败*\n\n"
	} else {
		prefix = "🏗️ *构建输出*\n\n"
	}
	
	return prefix + "```\n" + result + "\n```"
}

// TestCommand 测试命令
type TestCommand struct{}

func (t *TestCommand) Command() string {
	return "test"
}

func (t *TestCommand) Description() string {
	return "运行测试用例"
}

func (t *TestCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	var packagePath string
	if len(args) > 0 {
		packagePath = args[0]
	}
	
	// 构建参数
	params := map[string]interface{}{}
	if packagePath != "" {
		params["package"] = packagePath
	}
	
	// 执行测试操作
	result, err := executeTool("test_run", params)
	if err != nil {
		return fmt.Sprintf("❌ 测试运行失败: %v", err)
	}
	
	return fmt.Sprintf("🧪 *测试结果*\n\n```\n%s\n```", result)
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
	// 执行重启操作
	result, err := executeTool("project_reboot", map[string]interface{}{})
	if err != nil {
		return fmt.Sprintf("❌ 重启失败: %v", err)
	}
	
	return fmt.Sprintf("🔄 *重启项目*\n\n%s", result)
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
	result, err := executeTool("dependency_check", map[string]interface{}{})
	if err != nil {
		return fmt.Sprintf("❌ 依赖检查失败: %v", err)
	}
	
	return fmt.Sprintf("📦 *依赖分析*\n\n```\n%s\n```", result)
}

// LogViewCommand 日志查看命令
type LogViewCommand struct{}

func (l *LogViewCommand) Command() string {
	return "logs"
}

func (l *LogViewCommand) Description() string {
	return "查看应用日志"
}

func (l *LogViewCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	// 执行日志查看操作
	result, err := executeTool("log_view", map[string]interface{}{
		"log_file": "./log/orange-agent.log",
		"lines":    50,
	})
	if err != nil {
		return fmt.Sprintf("❌ 查看日志失败: %v", err)
	}
	
	return fmt.Sprintf("📋 *应用日志*\n\n```\n%s\n```", result)
}