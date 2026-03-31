package client

import (
	"context"
	"fmt"
	"orange-agent/agent/interfaces"
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
	return resp.Choices[0].Content
}

func (c *client) call(ctx context.Context, message []llms.MessageContent) (*llms.ContentResponse, error) {
	resp, err := c.llm.GenerateContent(ctx, message)
	if err != nil {
		return nil, err
	}
	c.manager.SaveCallRecord(message, resp, c.AgentConfig)
	return resp, nil
}
