package command

import (
	"context"
	"fmt"
	"orange-agent/agent/tools/git"
	"orange-agent/domain"
	"strings"

	"gopkg.in/telebot.v3"
)

// GitStatusCommand Git状态命令
type GitStatusCommand struct{}

func (g *GitStatusCommand) Command() string {
	return "git"
}

func (g *GitStatusCommand) Description() string {
	return "查看Git状态和更改"
}

func (g *GitStatusCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	// 执行git diff操作
	result, err := git.GitDiff("", false)
	if err != nil {
		return fmt.Sprintf("❌ 获取Git状态失败: %v", err)
	}
	
	var response strings.Builder
	response.WriteString("📊 *Git 状态*\n\n")
	
	if strings.TrimSpace(result) == "" {
		response.WriteString("✅ 工作区干净，没有未提交的更改")
	} else {
		response.WriteString("📝 *未提交的更改:*\n\n")
		response.WriteString("```diff\n")
		response.WriteString(result)
		response.WriteString("\n```")
	}
	
	response.WriteString("\n\n📋 *可用Git命令:*\n")
	response.WriteString("• `/git` - 查看当前状态\n")
	response.WriteString("• `/commit <消息>` - 提交更改\n")
	response.WriteString("• `/push [分支]` - 推送到远程\n")
	
	return response.String()
}

// GitCommitCommand Git提交命令
type GitCommitCommand struct{}

func (g *GitCommitCommand) Command() string {
	return "commit"
}

func (g *GitCommitCommand) Description() string {
	return "提交Git更改"
}

func (g *GitCommitCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	if len(args) == 0 {
		return "❌ 请提供提交信息\n📝 用法: `/commit <提交信息>`\n示例: `/commit 修复bug`"
	}
	
	commitMessage := strings.Join(args, " ")
	
	// 执行git commit操作
	result, err := git.GitCommit(commitMessage)
	if err != nil {
		return fmt.Sprintf("❌ 提交失败: %v", err)
	}
	
	return fmt.Sprintf("✅ *提交成功*\n\n%s", result)
}

// GitPushCommand Git推送命令
type GitPushCommand struct{}

func (g *GitPushCommand) Command() string {
	return "push"
}

func (g *GitPushCommand) Description() string {
	return "推送到远程仓库"
}

func (g *GitPushCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	var branch string
	if len(args) > 0 {
		branch = args[0]
	}
	
	// 执行git push操作
	result, err := git.GitPush(branch, false)
	if err != nil {
		return fmt.Sprintf("❌ 推送失败: %v", err)
	}
	
	var response strings.Builder
	response.WriteString("✅ *推送成功*\n\n")
	if branch != "" {
		response.WriteString(fmt.Sprintf("分支: %s\n\n", branch))
	}
	response.WriteString(result)
	
	return response.String()
}