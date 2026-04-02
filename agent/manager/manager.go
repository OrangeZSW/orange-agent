package manager

import (
	"fmt"
	"orange-agent/agent/interfaces"
	"orange-agent/agent/tools/skill"
	"orange-agent/domain"
	"orange-agent/repository"
	"orange-agent/repository/resource"
	"orange-agent/utils/logger"
	"sync"

	"github.com/tmc/langchaingo/llms"
)

var (
	once         sync.Once
	telegramInst interfaces.Telegram
)

type manager struct {
	user     *domain.User
	log      *logger.Logger
	repo     *repository.Repositories
	telegram interfaces.Telegram
}

// NewManager 创建管理器
func NewManager(user *domain.User) interfaces.Manager {
	return &manager{
		log:      logger.GetLogger(),
		repo:     resource.GetRepositories(),
		user:     user,
		telegram: telegramInst,
	}
}

// SetTelegram 设置Telegram实例
func SetTelegram(tg interfaces.Telegram) {
	once.Do(func() {
		telegramInst = tg
	})
}

// SaveCallRecord 保存调用记录
func (m *manager) SaveCallRecord(messages []llms.MessageContent, resp *llms.ContentResponse, agentConfig *domain.AgentConfig) error {
	memories, err := m.repo.Memory.GetMemoryByUserIdAndLimit(m.user.ID, 1)
	if err != nil || len(memories) == 0 {
		return fmt.Errorf("获取记忆失败")
	}

	record := &domain.CallRecord{
		MemoryId:         memories[0].ID,
		UserID:           m.user.ID,
		AgentId:          agentConfig.ID,
		ModelName:        m.user.ModelName,
		PromptTokens:     resp.Choices[0].GenerationInfo["PromptTokens"].(int),
		CompletionTokens: resp.Choices[0].GenerationInfo["CompletionTokens"].(int),
		TotalTokens:      resp.Choices[0].GenerationInfo["TotalTokens"].(int),
	}
	return m.repo.AgentCallRecord.CreateAgentCallRecord(record)
}

// SendMessage 通过Telegram发送消息
func (m *manager) SendMessage(text string) {
	if m.telegram != nil {
		m.telegram.SendMessage(int64(m.user.TelegramId), text)
	}
}

// SystemPrompt 获取系统提示词
func (m *manager) SystemPrompt() []llms.MessageContent {
	skills := skill.GetSkills()
	prompt := fmt.Sprintf("当用户提到的问题中包含以下技能，使用工具读取技能信息，根据技能完成任务\n所有的技能:%v", skills)
	return []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, prompt),
	}
}
