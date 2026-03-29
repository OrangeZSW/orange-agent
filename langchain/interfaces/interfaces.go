package interfaces

import (
	"context"

	"orange-agent/domain"

	"github.com/tmc/langchaingo/llms"
)

type LLMProvider interface {
	GetLLM(model string) (interface{}, error)
	GetDefaultModelName() string
	Call(ctx context.Context, messages []llms.MessageContent, tools []llms.Tool) (*llms.ContentResponse, error)
}

type MemoryManager interface {
	GetMemory(userID uint, limit int) ([]*domain.Memory, error)
	SaveMemory(memory *domain.Memory) error
}

type MessageBuilder interface {
	BuildMessages(user *domain.User, question string, prompt string) []llms.MessageContent
	BuildToolMessages(ctx context.Context, toolCalls []llms.ToolCall, messages []llms.MessageContent) ([]llms.MessageContent, error)
}

type MessageCleaner interface {
	CleanByToken(messages []llms.MessageContent, maxTokens int) ([]llms.MessageContent, error)
	CleanByCount(messages []llms.MessageContent, maxMessages int) []llms.MessageContent
}

type TokenCounter interface {
	CalculateTokens(msg llms.MessageContent) (int, error)
	MessageToText(msg llms.MessageContent) string
}

type ToolExecutor interface {
	Execute(ctx context.Context, toolName string, arguments string) (string, error)
	GetAvailableTools() map[string]interface{}
}

type ToolManager interface {
	HandleToolCalls(ctx context.Context, messages []llms.MessageContent, response *llms.ContentResponse, llm LLMProvider) (*llms.ContentResponse, error)
}

type CallRecordSaver interface {
	SaveCallRecord(user *domain.User, response *llms.ContentResponse, agentID uint, memoryID uint) error
}

type Chain interface {
	Process(ctx context.Context, user *domain.User, question string, prompt string) (string, error)
}
