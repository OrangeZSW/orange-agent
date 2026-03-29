package chain

import (
	"context"
	"fmt"

	"orange-agent/domain"
	"orange-agent/langchain/llm"
	"orange-agent/langchain/memory"
	"orange-agent/langchain/message"
	"orange-agent/langchain/tool"
	"orange-agent/tools"
	"orange-agent/utils/logger"
)

type Chain struct {
	llmProvider    *llm.OpenAIProvider
	memoryManager  memory.Manager
	messageBuilder *message.Builder
	toolManager    *tool.Manager
	log            *logger.Logger
}

func NewChain() *Chain {
	tokenCounter := message.NewTokenCounter()
	messageCleaner := message.NewCleaner(tokenCounter)
	memoryManager := memory.NewDBMemoryManager()
	messageBuilder := message.NewBuilder(memoryManager)
	toolExecutor := tool.NewExecutor()
	toolManager := tool.NewManager(toolExecutor, messageCleaner)
	llmProvider := llm.NewOpenAIProvider()

	return &Chain{
		llmProvider:    llmProvider,
		memoryManager:  memoryManager,
		messageBuilder: messageBuilder,
		toolManager:    toolManager,
		log:            logger.GetLogger(),
	}
}

func (c *Chain) Process(ctx context.Context, user *domain.User, memoryID uint, question string, prompt string) (string, error) {
	_, err := c.llmProvider.GetLLM(user.ModelName)
	if err != nil {
		c.log.Error("获取 LLM 失败: %v", err)
		return "", fmt.Errorf("获取 LLM 失败: %w", err)
	}
	c.toolManager.SetUser(user)
	c.llmProvider.SetContext(user, memoryID)

	c.log.Info("准备调用模型[%s][%s]", c.llmProvider.GetCurrentConfig().Name, user.ModelName)

	messages := c.messageBuilder.BuildMessages(user, question, prompt)

	toolList := tools.GetEllTools()
	response, err := c.llmProvider.Call(ctx, messages, toolList)
	if err != nil {
		c.log.Error("调用语言模型失败: %v", err)
		return "", fmt.Errorf("调用语言模型失败: %w", err)
	}

	if len(response.Choices) == 0 || response.Choices[0] == nil {
		c.log.Warn("模型返回空选择列表")
		return "抱歉，我没有收到有效的回复", nil
	}

	choice := response.Choices[0]

	if len(choice.ToolCalls) > 0 {
		response, err = c.toolManager.HandleToolCalls(ctx, messages, response, c.llmProvider)
		if err != nil {
			c.log.Error("处理工具调用失败: %v", err)
			return "", fmt.Errorf("处理工具调用失败: %w", err)
		}
	}

	if len(response.Choices) == 0 || response.Choices[0] == nil {
		return "抱歉，处理过程中出现错误", nil
	}

	return response.Choices[0].Content, nil
}

func (c *Chain) GetDefaultModelName() string {
	return c.llmProvider.GetDefaultModelName()
}
