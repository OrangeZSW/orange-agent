package langchain

import (
	"context"
	"fmt"
	"orange-agent/domain"

	"github.com/pkoukk/tiktoken-go"
	"github.com/tmc/langchaingo/llms"
)

type ToolMessageManager struct {
	maxTokens int // 最大 token 数
	tokenizer *tiktoken.Tiktoken
}

func NewToolMessageManager(maxTokens int) *ToolMessageManager {
	tkm, err := tiktoken.GetEncoding("cl100k_base") // OpenAI 的编码
	if err != nil {
		return nil
	}

	return &ToolMessageManager{
		maxTokens: maxTokens,
		tokenizer: tkm,
	}
}

// buildToolMessages 构建包含工具调用和响应的消息
func (h *AnswerHandler) buildToolMessages(ctx context.Context, toolCalls []llms.ToolCall,
	messages []llms.MessageContent) ([]llms.MessageContent, error) {

	// AI 消息包含所有工具调用
	aiMessage := llms.MessageContent{
		Role:  llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{},
	}

	// 用于存储所有工具响应消息
	var toolMessages []llms.MessageContent

	// 执行每个工具调用
	for _, toolCall := range toolCalls {
		// 添加AI的工具调用信息
		aiMessage.Parts = append(aiMessage.Parts, llms.ToolCall{
			ID:   toolCall.ID,
			Type: toolCall.Type,
			FunctionCall: &llms.FunctionCall{
				Name:      toolCall.FunctionCall.Name,
				Arguments: toolCall.FunctionCall.Arguments,
			},
		})

		h.logger.Info("执行工具调用：%s，参数：%.200s",
			toolCall.FunctionCall.Name,
			toolCall.FunctionCall.Arguments)

		// 执行工具
		result, err := h.executeTool(ctx, toolCall.FunctionCall.Name, toolCall.FunctionCall.Arguments)

		// 为每个工具调用创建独立的 tool 消息
		toolMessage := llms.MessageContent{
			Role:  llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{},
		}

		if err != nil {
			h.logger.Error("执行工具 %s 失败: %v", toolCall.FunctionCall.Name, err)
			toolMessage.Parts = append(toolMessage.Parts, llms.ToolCallResponse{
				ToolCallID: toolCall.ID,
				Content:    fmt.Sprintf("工具执行失败: %v", err),
				Name:       toolCall.FunctionCall.Name,
			})
		} else {
			h.logger.Info("工具调用 %s 成功，结果：%.50s", toolCall.FunctionCall.Name, result)
			toolMessage.Parts = append(toolMessage.Parts, llms.ToolCallResponse{
				ToolCallID: toolCall.ID,
				Content:    result,
				Name:       toolCall.FunctionCall.Name,
			})
		}

		// 添加到工具消息列表
		toolMessages = append(toolMessages, toolMessage)
	}

	h.logger.Info("工具调用处理完成，共处理 %d 个工具调用", len(toolCalls))

	// 先构建完整消息列表
	updatedMessages := make([]llms.MessageContent, 0, len(messages)+1+len(toolMessages))
	updatedMessages = append(updatedMessages, messages...)
	updatedMessages = append(updatedMessages, aiMessage)
	updatedMessages = append(updatedMessages, toolMessages...)

	// 基于 token 数量清理（使用配置的 maxTokens）
	cleanedMessages, err := h.cleanMessagesByToken(updatedMessages, 8000)
	if err != nil {
		h.logger.Error("清理消息失败: %v", err)
		// 降级：基于数量清理，保留最新的20条消息
		return h.cleanMessagesByCount(updatedMessages, 20), nil
	}

	return cleanedMessages, nil
}

// cleanMessagesByToken 基于 token 数量清理消息 (修复版：确保成对删除)
func (h *AnswerHandler) cleanMessagesByToken(
	messages []llms.MessageContent,
	maxTokens int,
) ([]llms.MessageContent, error) {

	if len(messages) == 0 {
		return messages, nil
	}

	// 1. 计算每条消息的 token 数
	totalTokens := 0
	tokenCounts := make([]int, len(messages))
	for i, msg := range messages {
		msgTokens, err := h.calculateTokens(msg)
		if err != nil {
			return nil, fmt.Errorf("计算消息 %d 的 token 数失败：%w", i, err)
		}
		tokenCounts[i] = msgTokens
		totalTokens += msgTokens
	}

	// 如果未超限，直接返回
	if totalTokens <= maxTokens {
		return messages, nil
	}

	// 2. 创建删除标记
	deleteFlags := make([]bool, len(messages))
	removedTokens := 0

	// 3. 策略：从最旧的消息开始扫描，寻找孤立的或需要被清理的 Tool 消息
	// 我们必须保持 [AI(tool_call), Tool(response)] 的完整性。
	// 如果我们要删除一个 Tool 消息，必须连同它前面的那个 AI 消息一起删除。

	// 我们从索引 1 开始遍历（因为 Tool 消息前面必须有 AI 消息，索引 0 不可能是待删除的孤立 Tool）
	for i := 1; i < len(messages); i++ {
		if totalTokens-removedTokens <= maxTokens {
			break
		}

		currentMsg := messages[i]

		// 只有当前消息是 Tool 消息时，才考虑删除这对消息
		if currentMsg.Role == llms.ChatMessageTypeTool {
			prevMsg := messages[i-1]

			// 检查前一条消息是否是包含 tool_calls 的 AI 消息
			// 注意：有些实现中，AI 消息可能混合了文本和 tool_call，只要角色是 AI 且紧接着是 Tool，通常视为一对
			if prevMsg.Role == llms.ChatMessageTypeAI {
				// 确认前一条 AI 消息确实包含工具调用（可选优化，防止误删普通 AI 回复）
				hasToolCall := false
				for _, part := range prevMsg.Parts {
					if _, ok := part.(llms.ToolCall); ok {
						hasToolCall = true
						break
					}
				}

				if hasToolCall {
					// === 核心修复：同时标记删除 AI 和 Tool ===

					// 如果前一条还没被标记删除（避免重复计算）
					if !deleteFlags[i-1] {
						deleteFlags[i-1] = true
						removedTokens += tokenCounts[i-1]
						h.logger.Debug("成对删除：旧的 AI 工具调用消息，索引：%d, tokens: %d", i-1, tokenCounts[i-1])
					}

					// 标记删除当前的 Tool 消息
					deleteFlags[i] = true
					removedTokens += tokenCounts[i]
					h.logger.Debug("成对删除：旧的工具响应消息，索引：%d, tokens: %d", i, tokenCounts[i])

					// 跳过下一条检查（因为 i+1 可能是新的 User 消息，不需要特殊处理，循环会继续）
					continue
				}
			}
		}
	}

	// 4. 构建清理后的消息列表
	result := make([]llms.MessageContent, 0, len(messages))
	for i, msg := range messages {
		if !deleteFlags[i] {
			result = append(result, msg)
		}
	}

	finalTokens := totalTokens - removedTokens
	h.logger.Info("消息清理完成，原始:%d -> 清理后:%d, 删除了 %d 条消息 (含成对删除)",
		totalTokens, finalTokens, len(messages)-len(result))

	return result, nil
}

// cleanMessagesByCount 基于数量清理消息（降级方案）
func (h *AnswerHandler) cleanMessagesByCount(
	messages []llms.MessageContent,
	maxMessages int,
) []llms.MessageContent {

	if len(messages) <= maxMessages {
		return messages
	}

	// 保留最新的 maxMessages 条消息
	startIndex := len(messages) - maxMessages
	result := make([]llms.MessageContent, maxMessages)
	copy(result, messages[startIndex:])

	h.logger.Warn("使用降级方案：基于数量清理，从 %d 条消息保留最新的 %d 条",
		len(messages), maxMessages)

	return result
}

// calculateTokens 计算单条消息的 token 数
func (h *AnswerHandler) calculateTokens(msg llms.MessageContent) (int, error) {
	// 如果 tokenizer 未初始化，返回估算值
	if h.langChain.toolMessageManager.tokenizer == nil {
		// 降级：粗略估算（1个token约等于4个字符）
		text := h.messageToText(msg)
		return len(text) / 4, nil
	}

	// 将消息转换为文本
	text := h.messageToText(msg)

	// 使用 tiktoken 计算 token 数
	tokens := h.langChain.toolMessageManager.tokenizer.Encode(text, nil, nil)
	return len(tokens), nil
}

// messageToText 将消息转换为文本
func (h *AnswerHandler) messageToText(msg llms.MessageContent) string {
	var text string

	// 添加角色标识
	switch msg.Role {
	case llms.ChatMessageTypeSystem:
		text += "System: "
	case llms.ChatMessageTypeHuman:
		text += "Human: "
	case llms.ChatMessageTypeAI:
		text += "AI: "
	case llms.ChatMessageTypeTool:
		text += "Tool: "
	}

	// 处理消息内容
	for _, part := range msg.Parts {
		switch p := part.(type) {
		case llms.TextContent:
			text += p.Text
		case llms.ToolCall:
			if p.FunctionCall != nil {
				text += fmt.Sprintf("ToolCall[name=%s, args=%s] ",
					p.FunctionCall.Name, p.FunctionCall.Arguments)
			}
		case llms.ToolCallResponse:
			text += fmt.Sprintf("ToolResponse[name=%s, content=%s] ",
				p.Name, p.Content)
		}
	}

	return text
}

// buildMessages 构建完整的对话消息
func (h *AnswerHandler) buildMessages(user *domain.User, question string, prompt string) []llms.MessageContent {
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, prompt),
	}

	// 添加用户记忆
	memories, err := h.repo.MemoryRepo.GetMemoryByUserIdAndLimit(user.ID, 3)

	if err != nil {
		h.logger.Error("获取用户记忆失败: %v", err)
	} else {
		h.logger.Debug("加载用户记忆：%d 条", len(memories))
		for _, memory := range memories {
			messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, memory.UserQuestion))
			if memory.AgentAnswer != "" {
				messages = append(messages, llms.TextParts(llms.ChatMessageTypeAI, memory.AgentAnswer))
			}
		}
	}

	// 添加当前问题
	messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, question))

	h.logger.Debug("构建的消息数量：%d", len(messages))
	return messages
}
