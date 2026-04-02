package client

import (
	"context"
	"fmt"
	"orange-agent/agent/interfaces"
	"orange-agent/agent/tools"
	"orange-agent/agent/utils"
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
	AgentConfig *domain.AgentConfig
	manager     interfaces.Manager
	compressor  *utils.ContextCompressor
}

func NewClient(Manager interfaces.Manager) interfaces.Client {
	return &client{
		repo:    resource.GetRepositories(),
		log:     logger.GetLogger(),
		manager: Manager,
	}
}

func (c *client) getLLM(modelName string) {
	config, err := c.repo.AgentConfig.GetAgentConfigByModel(modelName)
	c.AgentConfig = config
	if err != nil {
		c.log.Error("获取模型配置失败: %v", err)
	}
	c.log.Info("provider:[%-10s] model:[%-10s]", config.Name, modelName)
	llm, err := openai.New(
		openai.WithToken(config.Token),
		openai.WithBaseURL(config.BaseUrl),
		openai.WithModel(modelName),
	)
	if err != nil {
		c.log.Error("创建LLM失败: %v", err)
	}
	c.llm = llm
	// 重置压缩器，使用新的LLM实例
	c.compressor = utils.NewContextCompressor(c.llm)
}

func (c *client) Chat(modelName string, message []llms.MessageContent) string {
	ctx := context.Background()
	c.getLLM(modelName)
	resp, err := c.call(ctx, message)
	if err != nil {
		return fmt.Sprintf("调用LLM失败:%v", err)
	}
	if len(resp.Choices[0].ToolCalls) > 0 {
		return c.HandleToolCalls(ctx, message, resp)
	}

	return resp.Choices[0].Content
}

func (c *client) call(ctx context.Context, message []llms.MessageContent) (*llms.ContentResponse, error) {
	// 上下文压缩（仅对非系统消息的历史进行压缩）
	if c.compressor == nil {
		c.compressor = utils.NewContextCompressor(c.llm)
	}
	message = c.compressor.CompressIfNeeded(ctx, message)

	// 检查是否已有系统提示词，避免重复添加
	hasSystemPrompt := false
	for _, msg := range message {
		if msg.Role == llms.ChatMessageTypeSystem {
			hasSystemPrompt = true
			break
		}
	}

	// 只在首次调用时添加系统提示词
	if !hasSystemPrompt {
		message = append(message, c.manager.SystemPrompt()...)
	}

	resp, err := c.llm.GenerateContent(ctx, message, llms.WithTools(tools.GetEllTools()))
	if err != nil {
		return nil, err
	}
	c.manager.SaveCallRecord(message, resp, c.AgentConfig)
	return resp, nil
}

func (c *client) HandleToolCalls(ctx context.Context, message []llms.MessageContent, resp *llms.ContentResponse) string {
	toolsMessage := llms.MessageContent{
		Role:  llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{},
	}
	aiMessage := llms.MessageContent{
		Role:  llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{},
	}

	toolcalls := resp.Choices[0].ToolCalls
	if len(toolcalls) > 0 {
		for _, toolcall := range toolcalls {
			aiMessage.Parts = append(aiMessage.Parts, llms.ToolCall{
				ID:   toolcall.ID,
				Type: toolcall.Type,
				FunctionCall: &llms.FunctionCall{
					Name:      toolcall.FunctionCall.Name,
					Arguments: toolcall.FunctionCall.Arguments,
				},
			})
			res, err := tools.GetTools()[toolcall.FunctionCall.Name].Call(ctx, toolcall.FunctionCall.Arguments)
			if err != nil {
				c.log.Error("调用工具:%s失败,参数:%.20s,错误:%.50s", toolcall.FunctionCall.Name, toolcall.FunctionCall.Arguments, err)
				res = "调用工具失败"
			}
			toolsMessage.Parts = append(toolsMessage.Parts, llms.ToolCallResponse{
				ToolCallID: toolcall.ID,
				Content:    res,
				Name:       toolcall.FunctionCall.Name,
			})
		}
	}
	message = append(message, aiMessage)
	message = append(message, toolsMessage)
	resp, err := c.call(ctx, message)
	if err != nil {
		return fmt.Sprintf("工具调用中-调用LLM失败:%v", err)
	}
	if len(resp.Choices[0].ToolCalls) > 0 {
		return c.HandleToolCalls(ctx, message, resp)
	}
	return resp.Choices[0].Content
}
