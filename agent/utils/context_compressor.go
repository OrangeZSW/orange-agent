package utils

import (
	"context"
	"fmt"
	"orange-agent/utils/logger"

	"github.com/tmc/langchaingo/llms"
)

const (
	// MaxContextChars 最大上下文字符数（超过则触发压缩）
	MaxContextChars = 100000
	// CompressThreshold 触压压缩的阈值
	CompressThreshold = 100000
	// KeepRecentMessages 保留的最近消息数（不压缩）
	KeepRecentMessages = 4
)

// ContextCompressor 上下文压缩器
type ContextCompressor struct {
	llm llms.Model
	log *logger.Logger
}

// NewContextCompressor 创建上下文压缩器
func NewContextCompressor(llm llms.Model) *ContextCompressor {
	return &ContextCompressor{
		llm: llm,
		log: logger.GetLogger(),
	}
}

// CompressIfNeeded 检查并在需要时压缩上下文
func (c *ContextCompressor) CompressIfNeeded(
	ctx context.Context,
	messages []llms.MessageContent,
) []llms.MessageContent {
	// 估算当前字符数
	totalChars := estimateChars(messages)

	if totalChars < CompressThreshold {
		return messages // 不需要压缩
	}

	c.log.Info("触发上下文压缩: %d 字符 -> 目标 < %d", totalChars, CompressThreshold)

	// 分离系统提示词和历史消息
	var systemMsg *llms.MessageContent
	var historyMsgs []llms.MessageContent

	for _, msg := range messages {
		if msg.Role == llms.ChatMessageTypeSystem {
			tmp := msg
			systemMsg = &tmp
		} else {
			historyMsgs = append(historyMsgs, msg)
		}
	}

	// 保留最近的 N 条消息，压缩更早的消息
	if len(historyMsgs) <= KeepRecentMessages {
		return messages
	}

	toCompress := historyMsgs[:len(historyMsgs)-KeepRecentMessages]
	keepRecent := historyMsgs[len(historyMsgs)-KeepRecentMessages:]

	// 使用 LLM 对旧消息进行摘要
	summary := c.summarizeMessages(ctx, toCompress)

	// 重建消息列表
	result := []llms.MessageContent{}
	if systemMsg != nil {
		result = append(result, *systemMsg)
	}

	// 添加摘要作为系统消息
	summaryMsg := llms.MessageContent{
		Role: llms.ChatMessageTypeSystem,
		Parts: []llms.ContentPart{
			llms.TextContent{Text: fmt.Sprintf("【历史对话摘要】\n%s", summary)},
		},
	}
	result = append(result, summaryMsg)
	result = append(result, keepRecent...)

	newChars := estimateChars(result)
	c.log.Info("压缩完成: %d 字符 -> %d 字符 (减少 %.1f%%)",
		totalChars, newChars, float64(totalChars-newChars)/float64(totalChars)*100)

	return result
}

// summarizeMessages 使用 LLM 生成对话摘要
func (c *ContextCompressor) summarizeMessages(
	ctx context.Context,
	messages []llms.MessageContent,
) string {
	prompt := `请将以下对话压缩为简洁的摘要，保留关键信息：
- 用户的核心需求
- 已执行的操作和结果
- 重要的结论和待办事项
- 关键的代码片段或配置

对话记录：

`

	for _, msg := range messages {
		role := "用户"
		if msg.Role == llms.ChatMessageTypeAI {
			role = "助手"
		} else if msg.Role == llms.ChatMessageTypeTool {
			role = "工具"
		}

		for _, part := range msg.Parts {
			switch v := part.(type) {
			case llms.TextContent:
				prompt += fmt.Sprintf("%s: %s\n", role, truncateString(v.Text, 500))
			case llms.ToolCallResponse:
				prompt += fmt.Sprintf("工具[%s]: %s\n", v.Name, truncateString(v.Content, 300))
			case llms.ToolCall:
				prompt += fmt.Sprintf("助手调用工具: %s\n", v.FunctionCall.Name)
			}
		}
		prompt += "\n"
	}

	// 调用 LLM 生成摘要
	resp, err := c.llm.GenerateContent(ctx, []llms.MessageContent{
		{Role: llms.ChatMessageTypeHuman, Parts: []llms.ContentPart{
			llms.TextContent{Text: prompt},
		}},
	})

	if err != nil {
		c.log.Error("生成上下文摘要失败: %v", err)
		return "[摘要生成失败，历史对话已省略]"
	}

	if len(resp.Choices) == 0 {
		return "[摘要生成失败，历史对话已省略]"
	}

	return resp.Choices[0].Content
}

// estimateChars 估算消息列表的字符数
func estimateChars(messages []llms.MessageContent) int {
	total := 0
	for _, msg := range messages {
		for _, part := range msg.Parts {
			switch v := part.(type) {
			case llms.TextContent:
				total += len(v.Text)
			case llms.ToolCallResponse:
				total += len(v.Content)
			case llms.ToolCall:
				total += len(v.FunctionCall.Name) + len(v.FunctionCall.Arguments)
			}
		}
	}
	return total
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
