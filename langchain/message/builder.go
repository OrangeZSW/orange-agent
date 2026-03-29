package message

import (
	"orange-agent/domain"
	"orange-agent/langchain/memory"

	"github.com/tmc/langchaingo/llms"
)

type Builder struct {
	memoryManager memory.Manager
}

func NewBuilder(memoryManager memory.Manager) *Builder {
	return &Builder{
		memoryManager: memoryManager,
	}
}

func (b *Builder) BuildMessages(user *domain.User, question string, prompt string) []llms.MessageContent {
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, prompt),
	}

	memories, err := b.memoryManager.GetMemory(user.ID, 3)
	if err == nil && len(memories) > 0 {
		for _, mem := range memories {
			messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, mem.UserQuestion))
			if mem.AgentAnswer != "" {
				messages = append(messages, llms.TextParts(llms.ChatMessageTypeAI, mem.AgentAnswer))
			}
		}
	}

	messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, question))

	return messages
}
