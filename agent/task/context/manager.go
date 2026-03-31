// context/manager.go
package context

import (
	"fmt"
	"orange-agent/domain"
	"sync"

	"github.com/pkoukk/tiktoken-go"
)

type ContextManager struct {
	mu        sync.RWMutex
	tokenizer *tiktoken.Tiktoken
	maxTokens int
	contexts  map[uint]*domain.TaskContext
}

func NewContextManager(maxTokens int) *ContextManager {
	// 使用 cl100k_base 编码器（适用于 GPT-4/GPT-3.5）
	tokenizer, _ := tiktoken.GetEncoding("cl100k_base")

	return &ContextManager{
		tokenizer: tokenizer,
		maxTokens: maxTokens,
		contexts:  make(map[uint]*domain.TaskContext),
	}
}

// CreateContext 为子任务创建独立上下文
func (cm *ContextManager) CreateContext(subTaskID uint, systemPrompt string) *domain.TaskContext {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ctx := &domain.TaskContext{
		SystemPrompt: systemPrompt,
		Messages:     []domain.Message{},
		TokenCount:   0,
		Metadata:     make(map[string]interface{}),
	}

	cm.contexts[subTaskID] = ctx
	return ctx
}

// GetContext 获取子任务的上下文
func (cm *ContextManager) GetContext(subTaskID uint) (*domain.TaskContext, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	ctx, exists := cm.contexts[subTaskID]
	if !exists {
		return nil, fmt.Errorf("context not found for subtask: %s", subTaskID)
	}
	return ctx, nil
}

// AddMessage 添加消息到上下文，自动管理Token
func (cm *ContextManager) AddMessage(subTaskID uint, role, content string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	ctx, exists := cm.contexts[subTaskID]
	if !exists {
		return fmt.Errorf("context not found")
	}

	msg := domain.Message{
		Role:    role,
		Content: content,
	}

	// 计算新消息的Token数
	msgTokens := cm.countTokens(content)

	// 检查是否会超过限制
	if ctx.TokenCount+msgTokens > cm.maxTokens {
		// 触发压缩策略
		if err := cm.compressContext(ctx); err != nil {
			return err
		}
	}

	ctx.Messages = append(ctx.Messages, msg)
	ctx.TokenCount += msgTokens

	return nil
}

// compressContext 压缩上下文，保留关键信息
func (cm *ContextManager) compressContext(ctx *domain.TaskContext) error {
	if len(ctx.Messages) <= 2 {
		return fmt.Errorf("cannot compress: too few messages")
	}

	// 策略：保留系统提示、最近5条消息，总结中间消息
	systemMsg := []domain.Message{}
	if len(ctx.Messages) > 0 && ctx.Messages[0].Role == "system" {
		systemMsg = ctx.Messages[:1]
		ctx.Messages = ctx.Messages[1:]
	}

	// 保留最后5条消息
	keepCount := 5
	if len(ctx.Messages) <= keepCount {
		return nil
	}

	// 总结中间消息
	toSummarize := ctx.Messages[:len(ctx.Messages)-keepCount]
	summary := cm.summarizeMessages(toSummarize)

	// 构建新消息列表
	newMessages := make([]domain.Message, 0)
	newMessages = append(newMessages, systemMsg...)
	newMessages = append(newMessages, domain.Message{
		Role:    "system",
		Content: fmt.Sprintf("Previous conversation summary: %s", summary),
	})
	newMessages = append(newMessages, ctx.Messages[len(ctx.Messages)-keepCount:]...)

	// 重新计算Token
	ctx.Messages = newMessages
	ctx.TokenCount = cm.calculateTotalTokens(ctx.Messages)

	return nil
}

// summarizeMessages 总结消息（调用LLM进行压缩）
func (cm *ContextManager) summarizeMessages(messages []domain.Message) string {
	// 这里应该调用LLM进行总结
	// 简化实现：拼接前100个字符
	summary := ""
	for _, msg := range messages {
		if len(msg.Content) > 100 {
			summary += msg.Content[:100] + "...\n"
		} else {
			summary += msg.Content + "\n"
		}
	}
	return summary
}

// countTokens 计算Token数
func (cm *ContextManager) countTokens(text string) int {
	tokens := cm.tokenizer.Encode(text, nil, nil)
	return len(tokens)
}

// calculateTotalTokens 计算总Token数
func (cm *ContextManager) calculateTotalTokens(messages []domain.Message) int {
	total := 0
	for _, msg := range messages {
		total += cm.countTokens(msg.Content)
		// 每条消息的元数据开销
		total += 4
	}
	return total
}

// ClearContext 清除上下文（任务完成后）
func (cm *ContextManager) ClearContext(subTaskID uint) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.contexts, subTaskID)
}
