package manager

import (
	"orange-agent/agent/interfaces"
	"orange-agent/domain"
	"orange-agent/repository"
	"orange-agent/repository/resource"
	"orange-agent/utils/logger"

	"github.com/tmc/langchaingo/llms"
)

type manager struct {
	User   *domain.User
	log    *logger.Logger
	repo   *repository.Repositories
	memory *domain.Memory
}

func NewManager(user *domain.User, memory *domain.Memory) interfaces.Manager {
	return &manager{
		log:    logger.GetLogger(),
		repo:   resource.GetRepositories(),
		User:   user,
		memory: memory,
	}
}

func (r *manager) SaveCallRecord(message []llms.MessageContent, resp *llms.ContentResponse, agentConfig *domain.AgentConfig) error {
	//获取当前用户问题
	callRecord := &domain.CallRecord{
		MemoryId:         r.memory.ID,
		UserID:           r.User.ID,
		AgentId:          agentConfig.ID,
		ModelName:        agentConfig.Name,
		PromptTokens:     resp.Choices[0].GenerationInfo["PromptTokens"].(int),
		CompletionTokens: resp.Choices[0].GenerationInfo["CompletionTokens"].(int),
		TotalTokens:      resp.Choices[0].GenerationInfo["TotalTokens"].(int),
	}
	return r.repo.AgentCallRecord.CreateAgentCallRecord(callRecord)
}
