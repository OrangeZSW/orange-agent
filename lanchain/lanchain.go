package lanchain

import (
	"orange-agent/domain"
	"orange-agent/repository/factory"
	"orange-agent/utils/logger"

	"github.com/tmc/langchaingo/llms/openai"
)

type Lnachain struct {
	repoFactory factory.Factory
	log         *logger.Logger
	agentConfig *domain.AgentConfig
}

func NewLnachain() *Lnachain {
	return &Lnachain{
		repoFactory: *factory.NewFactory(),
		log:         logger.GetLogger(),
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
