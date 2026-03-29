package message

import (
	"fmt"

	"orange-agent/utils/logger"

	"github.com/tmc/langchaingo/llms"
)

type Cleaner struct {
	tokenCounter *TokenCounter
	log          *logger.Logger
}

func NewCleaner(tokenCounter *TokenCounter) *Cleaner {
	return &Cleaner{
		tokenCounter: tokenCounter,
		log:          logger.GetLogger(),
	}
}

func (c *Cleaner) CleanByToken(messages []llms.MessageContent, maxTokens int) ([]llms.MessageContent, error) {
	if len(messages) == 0 {
		return messages, nil
	}

	totalTokens := 0
	tokenCounts := make([]int, len(messages))
	for i, msg := range messages {
		msgTokens, err := c.tokenCounter.CalculateTokens(msg)
		if err != nil {
			return nil, fmt.Errorf("计算消息 %d 的 token 数失败：%w", i, err)
		}
		tokenCounts[i] = msgTokens
		totalTokens += msgTokens
	}

	if totalTokens <= maxTokens {
		return messages, nil
	}

	deleteFlags := make([]bool, len(messages))
	removedTokens := 0

	for i := 1; i < len(messages); i++ {
		if totalTokens-removedTokens <= maxTokens {
			break
		}

		currentMsg := messages[i]

		if currentMsg.Role == llms.ChatMessageTypeTool {
			prevMsg := messages[i-1]

			if prevMsg.Role == llms.ChatMessageTypeAI {
				hasToolCall := false
				for _, part := range prevMsg.Parts {
					if _, ok := part.(llms.ToolCall); ok {
						hasToolCall = true
						break
					}
				}

				if hasToolCall {
					if !deleteFlags[i-1] {
						deleteFlags[i-1] = true
						removedTokens += tokenCounts[i-1]
						c.log.Debug("成对删除：旧的 AI 工具调用消息，索引：%d, tokens: %d", i-1, tokenCounts[i-1])
					}

					deleteFlags[i] = true
					removedTokens += tokenCounts[i]
					c.log.Debug("成对删除：旧的工具响应消息，索引：%d, tokens: %d", i, tokenCounts[i])

					continue
				}
			}
		}
	}

	result := make([]llms.MessageContent, 0, len(messages))
	for i, msg := range messages {
		if !deleteFlags[i] {
			result = append(result, msg)
		}
	}

	finalTokens := totalTokens - removedTokens
	c.log.Info("消息清理完成，原始:%d -> 清理后:%d, 删除了 %d 条消息",
		totalTokens, finalTokens, len(messages)-len(result))

	return result, nil
}

func (c *Cleaner) CleanByCount(messages []llms.MessageContent, maxMessages int) []llms.MessageContent {
	if len(messages) <= maxMessages {
		return messages
	}

	startIndex := len(messages) - maxMessages
	result := make([]llms.MessageContent, maxMessages)
	copy(result, messages[startIndex:])

	c.log.Warn("使用降级方案：基于数量清理，从 %d 条消息保留最新的 %d 条",
		len(messages), maxMessages)

	return result
}
