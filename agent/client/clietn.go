package client

import (
	"context"
	"fmt"
	"orange-agent/agent/interfaces"
	"orange-agent/agent/tools"
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
			c.log.Info("调用工具:%s,参数:%.20s", toolcall.FunctionCall.Name, toolcall.FunctionCall.Arguments)
			c.manager.TeleGramSendMessage(fmt.Sprintf("调用工具:%s,参数:%v", toolcall.FunctionCall.Name, toolcall.FunctionCall.Arguments))
			res, err := tools.GetTools()[toolcall.FunctionCall.Name].Call(ctx, toolcall.FunctionCall.Arguments)
			if err != nil {
				c.log.Error("调用工具:%s失败,参数:%.20s,错误:%.200s", toolcall.FunctionCall.Name, toolcall.FunctionCall.Arguments, err)
				res = "调用工具失败"
			}
			c.log.Info("调用工具:%s成功,参数:%.20s,工具输出:%.200s", toolcall.FunctionCall.Name, toolcall.FunctionCall.Arguments, res)
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
