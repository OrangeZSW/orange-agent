package utils

import (
	"fmt"
	"strings"
)

type ToolMessageFormatter struct{}

func NewToolMessageFormatter() *ToolMessageFormatter {
	return &ToolMessageFormatter{}
}

// FormatToolCallMessage 格式化工具调用消息
func (f *ToolMessageFormatter) FormatToolCallMessage(toolName, arguments string) string {
	// 美化工具名
	prettyToolName := f.prettifyToolName(toolName)

	// 格式化参数
	prettyArgs := f.prettifyArguments(arguments)

	return fmt.Sprintf("🛠️ *工具调用*\n\n📋 *工具名称*: %s\n⚙️ *参数*:\n%s", prettyToolName, prettyArgs)
}

// FormatToolSuccessMessage 格式化工具调用成功消息
func (f *ToolMessageFormatter) FormatToolSuccessMessage(toolName, arguments, result string) string {
	// 美化工具名
	prettyToolName := f.prettifyToolName(toolName)

	// 格式化参数
	prettyArgs := f.prettifyArguments(arguments)

	// 格式化结果（截断过长的结果）
	prettyResult := f.prettifyResult(result)

	return fmt.Sprintf("✅ *工具调用成功*\n\n📋 *工具名称*: %s\n⚙️ *参数*:\n%s\n📊 *输出*:\n%.50s",
		prettyToolName, prettyArgs, prettyResult)
}

// FormatToolErrorMessage 格式化工具调用失败消息
func (f *ToolMessageFormatter) FormatToolErrorMessage(toolName, arguments, error string) string {
	// 美化工具名
	prettyToolName := f.prettifyToolName(toolName)

	// 格式化参数
	prettyArgs := f.prettifyArguments(arguments)

	return fmt.Sprintf("❌ *工具调用失败*\n\n📋 *工具名称*: %s\n⚙️ *参数*:\n%s\n💥 *错误*:\n%s",
		prettyToolName, prettyArgs, error)
}

// prettifyToolName 美化工具名称
func (f *ToolMessageFormatter) prettifyToolName(toolName string) string {
	// 将下划线转换为空格并首字母大写
	parts := strings.Split(toolName, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, " ")
}

// prettifyArguments 美化参数显示
func (f *ToolMessageFormatter) prettifyArguments(args string) string {
	if args == "" || args == "{}" {
		return "`无参数`"
	}

	// 格式化JSON参数
	formatted := strings.ReplaceAll(args, ",", ",\n")
	formatted = strings.ReplaceAll(formatted, "{", "{\n")
	formatted = strings.ReplaceAll(formatted, "}", "\n}")

	// 添加代码块标记
	return fmt.Sprintf("```json\n%s\n```", formatted)
}

// prettifyResult 美化结果显示
func (f *ToolMessageFormatter) prettifyResult(result string) string {
	if result == "" {
		return "`无输出`"
	}

	// 如果结果太长，截断并添加省略号
	maxLength := 500
	if len(result) > maxLength {
		truncated := result[:maxLength] + "..."
		// 尝试保持JSON格式的完整性
		if strings.Contains(result, "{") {
			// 如果是JSON，确保结束括号
			if !strings.Contains(truncated, "}") {
				truncated += "\n...\n}"
			}
		}
		return fmt.Sprintf("```\n%s\n```", truncated)
	}

	// 尝试判断是否是JSON格式
	if strings.HasPrefix(strings.TrimSpace(result), "{") ||
		strings.HasPrefix(strings.TrimSpace(result), "[") {
		// 尝试格式化JSON
		result = strings.ReplaceAll(result, "\\n", "\n")
		result = strings.ReplaceAll(result, "\\\"", "\"")
		return fmt.Sprintf("```json\n%s\n```", result)
	}

	return fmt.Sprintf("```\n%s\n```", result)
}
