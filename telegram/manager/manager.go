package manager

import (
	"orange-agent/domain"
	"orange-agent/repository"
	"orange-agent/repository/resource"
	"orange-agent/telegram/interfaces"
	"orange-agent/utils"

	"github.com/tmc/langchaingo/llms"
)

type Manager struct {
	repo *repository.Repositories
}

func NewManager() interfaces.Manager {
	return &Manager{
		repo: resource.GetRepositories(),
	}
}

func (m *Manager) GetUser(telegramId int64, name string) *domain.User {
	user, _ := m.repo.User.GetUserByTelegramId(telegramId)
	if user == nil {
		user = &domain.User{
			TelegramId: utils.Int64ToUint(telegramId),
			Name:       name,
			ModelName:  "gpt-3.5-turbo",
		}
		m.repo.User.CreateUser(user)
	}
	return user
}

func (m *Manager) GetMessage(id uint, question string) []llms.MessageContent {
	message := []llms.MessageContent{}
	memorys, _ := m.repo.Memory.GetMemoryByUserIdAndLimit(id, 3)
	for _, item := range memorys {
		message = append(message, llms.TextParts(llms.ChatMessageTypeHuman, item.UserQuestion))
		message = append(message, llms.TextParts(llms.ChatMessageTypeAI, item.AgentAnswer))
	}
	message = append(message, llms.TextParts(llms.ChatMessageTypeHuman, question))
	return message
}
