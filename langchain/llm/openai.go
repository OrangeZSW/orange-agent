package llm

import (
	"context"
	"fmt"

	"orange-agent/domain"
	"orange-agent/repository/factory"
	"orange-agent/utils/logger"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type OpenAIProvider struct {
	repoFactory factory.Factory
	log         *logger.Logger
	agentConfig *domain.AgentConfig
	llm         *openai.LLM
}

func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{
		repoFactory: *factory.NewFactory(),
		log:         logger.GetLogger(),
	}
}

func (p *OpenAIProvider) GetLLM(model string) (*openai.LLM, error) {
	config, err := p.repoFactory.AgentConfigRepo.GetAgentConfigByModel(model)
	if err != nil {
		p.log.Error("获取模型 %s 配置失败: %v", model, err)
		return nil, fmt.Errorf("获取模型配置失败: %w", err)
	}

	p.agentConfig = config

	llm, err := openai.New(
		openai.WithModel(model),
		openai.WithBaseURL(config.BaseUrl),
		openai.WithToken(config.Token),
	)
	if err != nil {
		p.log.Error("创建 OpenAI LLM 失败: %v", err)
		return nil, fmt.Errorf("创建 LLM 失败: %w", err)
	}

	p.llm = llm
	return llm, nil
}

func (p *OpenAIProvider) GetDefaultModelName() string {
	config, err := p.repoFactory.AgentConfigRepo.GetAgentConfigByName("default")
	if err != nil {
		p.log.Error("获取默认模型名称失败: %v", err)
		return "qwen3:8b"
	}

	if config == nil || len(config.Models) == 0 {
		return "qwen3:8b"
	}

	return config.Models[0]
}

func (p *OpenAIProvider) Call(ctx context.Context, messages []llms.MessageContent, tools []llms.Tool) (*llms.ContentResponse, error) {
	if p.llm == nil {
		return nil, fmt.Errorf("LLM 未初始化")
	}

	response, err := p.llm.GenerateContent(ctx, messages, llms.WithTools(tools))
	if err != nil {
		return nil, fmt.Errorf("调用 LLM 失败: %w", err)
	}

	return response, nil
}

func (p *OpenAIProvider) GetCurrentConfig() *domain.AgentConfig {
	return p.agentConfig
}
