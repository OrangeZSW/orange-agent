package manager

import (
	"fmt"
	"orange-agent/agent/interfaces"
	"orange-agent/agent/tools/skill"
	"orange-agent/domain"
	"orange-agent/repository"
	"orange-agent/repository/resource"
	"orange-agent/telegram"
	"orange-agent/utils/logger"
	"sync"

	"github.com/tmc/langchaingo/llms"
)

var (
	once sync.Once
)

type manager struct {
	User     *domain.User
	log      *logger.Logger
	repo     *repository.Repositories
	telegram interfaces.Telegram
}

func NewManager(user *domain.User) interfaces.Manager {
	return &manager{
		log:      logger.GetLogger(),
		repo:     resource.GetRepositories(),
		User:     user,
		telegram: telegram.GetTelegram(),
	}
}

func (r *manager) SaveCallRecord(message []llms.MessageContent, resp *llms.ContentResponse, agentConfig *domain.AgentConfig) error {
	memory, _ := r.repo.Memory.GetMemoryByUserIdAndLimit(r.User.ID, 1)

	//获取当前用户问题
	callRecord := &domain.CallRecord{
		MemoryId:         memory[0].ID,
		UserID:           r.User.ID,
		AgentId:          agentConfig.ID,
		ModelName:        r.User.ModelName,
		PromptTokens:     resp.Choices[0].GenerationInfo["PromptTokens"].(int),
		CompletionTokens: resp.Choices[0].GenerationInfo["CompletionTokens"].(int),
		TotalTokens:      resp.Choices[0].GenerationInfo["TotalTokens"].(int),
	}
	return r.repo.AgentCallRecord.CreateAgentCallRecord(callRecord)
}

func (r *manager) TeleGramSendMessage(text string) {
	r.telegram.SendTeleGramMessage(int64(r.User.TelegramId), text)
}

func (r *manager) SystemPrompt() []llms.MessageContent {
	skills := skill.GetSkills()
	skillsPrompt := []llms.MessageContent{}
	skillsPrompt = append(skillsPrompt, llms.TextParts(
		llms.ChatMessageTypeSystem,
		"当用户提到的问题中包含以下技能，使用工具读取技能信息，根据技能完成任务", fmt.Sprintf("所有的技能:%v", skills),
	))
	return skillsPrompt
}
