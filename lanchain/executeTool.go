package langchain

import (
	"context"
	"fmt"
	"orange-agent/domain"
	"orange-agent/tools"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// handleToolCalls 递归处理工具调用
func (h *AnswerHandler) handleToolCalls(ctx context.Context, user *domain.User, messages []llms.MessageContent,
	response *llms.ContentResponse, llm *openai.LLM) *llms.ContentResponse {

	choice := response.Choices[0]
	if choice == nil || len(choice.ToolCalls) == 0 {
		return response
	}

	// 构建包含工具调用的消息
	updatedMessages, err := h.buildToolMessages(ctx, choice.ToolCalls, messages)
	if err != nil {
		h.logger.Error("构建工具消息失败: %v", err)
		return h.createErrorResponse(response, fmt.Sprintf("构建工具消息失败: %v", err))
	}

	// 再次调用模型
	newResponse, err := h.CallLLM(ctx, updatedMessages, llm, user)
	if err != nil {
		h.logger.Error("工具调用后再次调用模型失败: %v", err)
		return h.createErrorResponse(response, fmt.Sprintf("工具调用失败: %v", err))
	}

	// 递归处理可能的进一步工具调用
	return h.handleToolCalls(ctx, user, updatedMessages, newResponse, llm)
}

// executeTool 执行具体的工具
func (h *AnswerHandler) executeTool(ctx context.Context, toolName string, arguments string) (string, error) {
	availableTools := tools.GetTools()

	// 检查工具是否存在
	if _, exists := availableTools[toolName]; !exists {
		h.logger.Error("未找到工具：%s", toolName)
		return "", fmt.Errorf("工具 '%s' 不存在", toolName)
	}

	// 执行工具
	result, err := availableTools[toolName].Call(ctx, arguments)
	if err != nil {
		return "", fmt.Errorf("工具调用失败: %w", err)
	}

	return result, nil
}
