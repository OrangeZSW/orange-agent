package langchain

import (
	"orange-agent/domain"
	repo_factory "orange-agent/repository/factory"
	"orange-agent/utils/logger"

	"github.com/tmc/langchaingo/llms/openai"
)

type Lnachain struct {
	repoFactory        repo_factory.Factory
	log                *logger.Logger
	agentConfig        *domain.AgentConfig
	toolMessageManager *ToolMessageManager
}

func NewLnachain() *Lnachain {
	return &Lnachain{
		repoFactory:        *repo_factory.NewFactory(),
		log:                logger.GetLogger(),
		toolMessageManager: NewToolMessageManager(2000),
	}
}
func (l *Lnachain) GetLLM(model string) *openai.LLM {
	config, err := l.repoFactory.AgentConfigRepo.GetAgentConfigByModel(model)
	l.agentConfig = config
	if err != nil {
		l.log.Error("%s模型配置文件获取失败", model)
	}
	llm, err := openai.New(
		openai.WithModel(model),
		openai.WithBaseURL(config.BaseUrl),
		openai.WithToken(config.Token),
	)
	return llm
}

// getdfault model name
func (l *Lnachain) GetDefaultModelName() string {
	config, err := l.repoFactory.AgentConfigRepo.GetAgentConfigByName("default")
	if err != nil {
		l.log.Error("get default model name error: %v", err)
	}
	if config == nil {
		return "qwen3:8b"

	}
	return config.Models[0]
}
