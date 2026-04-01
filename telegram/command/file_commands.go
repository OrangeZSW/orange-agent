package command

import (
	"context"
	"fmt"
	"orange-agent/domain"
	"strings"

	"gopkg.in/telebot.v3"
)

// FileListCommand 列出文件命令
type FileListCommand struct{}

func (f *FileListCommand) Command() string {
	return "list"
}

func (f *FileListCommand) Description() string {
	return "列出当前目录下的所有文件"
}

func (f *FileListCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	// 执行文件列表操作
	result, err := executeTool("file_list", map[string]interface{}{})
	if err != nil {
		return fmt.Sprintf("❌ 获取文件列表失败: %v", err)
	}
	
	// 格式化结果
	var response strings.Builder
	response.WriteString("📁 *文件列表*\n\n")
	
	// 按行分割结果
	lines := strings.Split(strings.TrimSpace(result), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			response.WriteString(fmt.Sprintf("• %s\n", line))
		}
	}
	
	if len(lines) == 0 || (len(lines) == 1 && strings.TrimSpace(lines[0]) == "") {
		response.WriteString("当前目录为空")
	}
	
	return response.String()
}

// FileReadCommand 读取文件命令
type FileReadCommand struct{}

func (f *FileReadCommand) Command() string {
	return "read"
}

func (f *FileReadCommand) Description() string {
	return "读取指定文件内容"
}

func (f *FileReadCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	if len(args) == 0 {
		return "❌ 请指定要读取的文件路径\n📝 用法: `/read <文件路径>`\n示例: `/read main.go`"
	}
	
	filePath := args[0]
	
	// 执行文件读取操作
	result, err := executeTool("file_read", map[string]interface{}{
		"file_path": filePath,
	})
	if err != nil {
		return fmt.Sprintf("❌ 读取文件失败: %v", err)
	}
	
	// 如果内容太长，截断
	maxLength := 2000
	if len(result) > maxLength {
		result = result[:maxLength] + "\n\n... (内容过长，已截断)"
	}
	
	// 格式化结果
	var response strings.Builder
	response.WriteString(fmt.Sprintf("📄 *文件内容: %s*\n\n", filePath))
	response.WriteString("```\n")
	response.WriteString(result)
	response.WriteString("\n```")
	
	return response.String()
}

// FileSearchCommand 搜索文件命令
type FileSearchCommand struct{}

func (f *FileSearchCommand) Command() string {
	return "search"
}

func (f *FileSearchCommand) Description() string {
	return "搜索文件内容"
}

func (f *FileSearchCommand) Handle(ctx context.Context, c telebot.Context, user *domain.User, args []string) string {
	if len(args) == 0 {
		return "❌ 请指定要搜索的内容\n📝 用法: `/search <搜索内容>`\n示例: `/search function`"
	}
	
	searchPattern := strings.Join(args, " ")
	
	// 执行文件搜索操作
	result, err := executeTool("file_search", map[string]interface{}{
		"pattern": searchPattern,
	})
	if err != nil {
		return fmt.Sprintf("❌ 搜索文件失败: %v", err)
	}
	
	var response strings.Builder
	response.WriteString(fmt.Sprintf("🔍 *搜索结果: '%s'*\n\n", searchPattern))
	
	// 检查是否有结果
	if strings.TrimSpace(result) == "" {
		response.WriteString("未找到匹配的内容")
	} else {
		response.WriteString("```\n")
		response.WriteString(result)
		response.WriteString("\n```")
	}
	
	return response.String()
}