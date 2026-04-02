package manager

import (
	"orange-agent/domain"
	"orange-agent/repository"
	"orange-agent/repository/resource"
	"orange-agent/telegram/interfaces"
	"orange-agent/utils"

	"github.com/tmc/langchaingo/llms"
)

const defaultHistoryLimit = 3

type Manager struct {
	repo *repository.Repositories
}

// NewManager 创建用户管理器
func NewManager() interfaces.Manager {
	return &Manager{
		repo: resource.GetRepositories(),
	}
}

// GetUser 获取或创建用户
func (m *Manager) GetUser(telegramId int64, name string) *domain.User {
	user, err := m.repo.User.GetUserByTelegramId(telegramId)
	if err != nil || user == nil {
		user = &domain.User{
			TelegramId: utils.Int64ToUint(telegramId),
			Name:       name,
			ModelName:  "gpt-3.5-turbo",
		}
		m.repo.User.CreateUser(user)
	}
	return user
}

// GetMessageHistory 获取用户的消息历史
func (m *Manager) GetMessageHistory(userId uint, limit int) []llms.MessageContent {
	if limit <= 0 {
		limit = defaultHistoryLimit
	}

	messages := []llms.MessageContent{}
	memories, err := m.repo.Memory.GetMemoryByUserIdAndLimit(userId, limit)
	if err != nil {
		return messages
	}

	for _, item := range memories {
		messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, item.UserQuestion))
		messages = append(messages, llms.TextParts(llms.ChatMessageTypeAI, item.AgentAnswer))
	}
	return messages
}
