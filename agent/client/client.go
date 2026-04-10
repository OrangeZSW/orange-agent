package client

import (
	"context"
	"fmt"
	"orange-agent/agent/interfaces"
	"orange-agent/agent/tools"
	agentutils "orange-agent/agent/utils"
	"orange-agent/domain"
	"orange-agent/repository"
	"orange-agent/repository/resource"
	"orange-agent/utils/logger"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type client struct {
	llm         *openai.LLM
	repo        *repository.Repositories
	log         *logger.Logger
	agentConfig *domain.AgentConfig
	manager     interfaces.Manager
	compressor  *agentutils.ContextCompressor
}

// NewClient 创建Agent客户端
func NewClient(mgr interfaces.Manager) interfaces.Client {
	return &client{
		repo:    resource.GetRepositories(),
		log:     logger.GetLogger(),
		manager: mgr,
	}
}

// getLLM 获取LLM实例
func (c *client) getLLM(modelName string) error {
	config, err := c.repo.AgentConfig.GetAgentConfigByModel(modelName)
	if err != nil {
		c.log.Error("获取模型配置失败: %v", err)
		return err
	}
	c.agentConfig = config
	c.log.Info("provider:[%-10s] model:[%-10s]", config.Name, modelName)

	llm, err := openai.New(
		openai.WithToken(config.Token),
		openai.WithBaseURL(config.BaseUrl),
		openai.WithModel(modelName),
	)
	if err != nil {
		c.log.Error("创建LLM失败: %v", err)
		return err
	}
	c.llm = llm
	c.compressor = agentutils.NewContextCompressor(c.llm)
	return nil
}

// Chat 与AI模型对话
func (c *client) Chat(modelName string, messages []llms.MessageContent) string {
	if err := c.getLLM(modelName); err != nil {
		return fmt.Sprintf("初始化LLM失败: %v", err)
	}

	ctx := context.Background()
	resp, err := c.call(ctx, messages)
	if err != nil {
		return fmt.Sprintf("调用LLM失败: %v", err)
	}

	if len(resp.Choices[0].ToolCalls) > 0 {
		return c.handleToolCalls(ctx, messages, resp)
	}
	return resp.Choices[0].Content
}

// call 调用LLM
func (c *client) call(ctx context.Context, messages []llms.MessageContent) (*llms.ContentResponse, error) {
	// 检查是否已有系统提示词，避免重复添加
	hasSystemPrompt := false
	for _, msg := range messages {
		if msg.Role == llms.ChatMessageTypeSystem {
			hasSystemPrompt = true
			break
		}
	}
	// 只在首次调用时添加系统提示词
	if !hasSystemPrompt {
		messages = append(messages, c.manager.SystemPrompt()...)
	}

	resp, err := c.llm.GenerateContent(ctx, messages, llms.WithTools(tools.GetEllTools()))
	if err != nil {
		return nil, err
	}

	c.manager.SaveCallRecord(messages, resp, c.agentConfig)
	return resp, nil
}

// handleToolCalls 处理工具调用
func (c *client) handleToolCalls(ctx context.Context, messages []llms.MessageContent, resp *llms.ContentResponse) string {
	toolMessages := llms.MessageContent{
		Role:  llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{},
	}
	aiMessage := llms.MessageContent{
		Role:  llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{},
	}

	for _, toolCall := range resp.Choices[0].ToolCalls {
		aiMessage.Parts = append(aiMessage.Parts, llms.ToolCall{
			ID:   toolCall.ID,
			Type: toolCall.Type,
			FunctionCall: &llms.FunctionCall{
				Name:      toolCall.FunctionCall.Name,
				Arguments: toolCall.FunctionCall.Arguments,
			},
		})

		// 执行工具
		c.log.Info("调用工具:%s,参数:```%.20s```", toolCall.FunctionCall.Name, toolCall.FunctionCall.Arguments)
		result, err := tools.GetTools()[toolCall.FunctionCall.Name].Call(ctx, toolCall.FunctionCall.Arguments)
		if err != nil {
			c.log.Error("调用工具:%s失败,参数:%.20s,错误:%.50s", toolCall.FunctionCall.Name, toolCall.FunctionCall.Arguments, err)
			result = "调用工具失败"
		}

		toolMessages.Parts = append(toolMessages.Parts, llms.ToolCallResponse{
			ToolCallID: toolCall.ID,
			Content:    result,
			Name:       toolCall.FunctionCall.Name,
		})
	}

	messages = append(messages, aiMessage, toolMessages)

	resp, err := c.call(ctx, messages)
	if err != nil {
		return fmt.Sprintf("工具调用中-调用LLM失败: %v", err)
	}

	if len(resp.Choices[0].ToolCalls) > 0 {
		return c.handleToolCalls(ctx, messages, resp)
	}
	return resp.Choices[0].Content
}
