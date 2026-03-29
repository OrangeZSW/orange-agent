package langchain

import (
	"context"
	"fmt"
	"orange-agent/domain"

	"github.com/tmc/langchaingo/llms"
)

// / buildToolMessages 构建包含工具调用和响应的消息
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

	// 构建新的消息列表
	updatedMessages := make([]llms.MessageContent, 0, len(messages)+1+len(toolMessages))
	updatedMessages = append(updatedMessages, messages...)
	updatedMessages = append(updatedMessages, aiMessage)
	updatedMessages = append(updatedMessages, toolMessages...)

	return updatedMessages, nil
}

// buildMessages 构建完整的对话消息
func (h *AnswerHandler) buildMessages(user *domain.User, question string, prompt string) []llms.MessageContent {
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, prompt),
	}

	// 添加用户记忆
	memories, err := h.repo.MemoryRepo.GetMemoryByIdAndSize(user.ID, 3)
	for _, item := range memories {
		item.AgentAnswer = ""
	}

	if err != nil {
		h.logger.Error("获取用户记忆失败: %v", err)
	} else {
		h.logger.Debug("加载用户记忆：%d 条", len(memories))
	}

	// 添加当前问题
	messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, question))

	h.logger.Debug("构建的消息数量：%d", len(messages))
	return messages
}
