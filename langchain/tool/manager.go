package tool

import (
	"context"
	"fmt"

	"orange-agent/langchain/llm"
	"orange-agent/langchain/message"
	"orange-agent/tools"
	"orange-agent/utils/logger"

	"github.com/tmc/langchaingo/llms"
)

type Manager struct {
	executor       *Executor
	messageCleaner *message.Cleaner
	log            *logger.Logger
}

func NewManager(executor *Executor, cleaner *message.Cleaner) *Manager {
	return &Manager{
		executor:       executor,
		messageCleaner: cleaner,
		log:            logger.GetLogger(),
	}
}

func (m *Manager) HandleToolCalls(ctx context.Context, messages []llms.MessageContent,
	response *llms.ContentResponse, llmProvider *llm.OpenAIProvider) (*llms.ContentResponse, error) {

	choice := response.Choices[0]
	if choice == nil || len(choice.ToolCalls) == 0 {
		return response, nil
	}

	updatedMessages, err := m.buildToolMessages(ctx, choice.ToolCalls, messages)
	if err != nil {
		m.log.Error("构建工具消息失败: %v", err)
		return nil, fmt.Errorf("构建工具消息失败: %w", err)
	}

	toolList := tools.GetEllTools()
	newResponse, err := llmProvider.Call(ctx, updatedMessages, toolList)
	if err != nil {
		m.log.Error("工具调用后再次调用模型失败: %v", err)
		return nil, fmt.Errorf("工具调用失败: %w", err)
	}

	return m.HandleToolCalls(ctx, updatedMessages, newResponse, llmProvider)
}

func (m *Manager) buildToolMessages(ctx context.Context, toolCalls []llms.ToolCall,
	messages []llms.MessageContent) ([]llms.MessageContent, error) {

	aiMessage := llms.MessageContent{
		Role:  llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{},
	}

	var toolMessages []llms.MessageContent

	for _, toolCall := range toolCalls {
		aiMessage.Parts = append(aiMessage.Parts, llms.ToolCall{
			ID:   toolCall.ID,
			Type: toolCall.Type,
			FunctionCall: &llms.FunctionCall{
				Name:      toolCall.FunctionCall.Name,
				Arguments: toolCall.FunctionCall.Arguments,
			},
		})

		m.log.Info("执行工具调用：%s，参数：%.200s",
			toolCall.FunctionCall.Name,
			toolCall.FunctionCall.Arguments)

		result, err := m.executor.Execute(ctx, toolCall.FunctionCall.Name, toolCall.FunctionCall.Arguments)

		toolMessage := llms.MessageContent{
			Role:  llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{},
		}

		if err != nil {
			m.log.Error("执行工具 %s 失败: %v", toolCall.FunctionCall.Name, err)
			toolMessage.Parts = append(toolMessage.Parts, llms.ToolCallResponse{
				ToolCallID: toolCall.ID,
				Content:    fmt.Sprintf("工具执行失败: %v", err),
				Name:       toolCall.FunctionCall.Name,
			})
		} else {
			m.log.Info("工具调用 %s 成功，结果：%.50s", toolCall.FunctionCall.Name, result)
			toolMessage.Parts = append(toolMessage.Parts, llms.ToolCallResponse{
				ToolCallID: toolCall.ID,
				Content:    result,
				Name:       toolCall.FunctionCall.Name,
			})
		}

		toolMessages = append(toolMessages, toolMessage)
	}

	m.log.Info("工具调用处理完成，共处理 %d 个工具调用", len(toolCalls))

	updatedMessages := make([]llms.MessageContent, 0, len(messages)+1+len(toolMessages))
	updatedMessages = append(updatedMessages, messages...)
	updatedMessages = append(updatedMessages, aiMessage)
	updatedMessages = append(updatedMessages, toolMessages...)

	cleanedMessages, err := m.messageCleaner.CleanByToken(updatedMessages, 8000)
	if err != nil {
		m.log.Error("清理消息失败: %v", err)
		return m.messageCleaner.CleanByCount(updatedMessages, 20), nil
	}

	return cleanedMessages, nil
}
