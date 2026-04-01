package agent

import (
	"context"
	"orange-agent/agent/client"
	"orange-agent/agent/interfaces"
	"orange-agent/agent/manager"
	"orange-agent/domain"
	"orange-agent/repository"
	"orange-agent/repository/resource"
	"orange-agent/utils"
	"sync"

	"github.com/tmc/langchaingo/llms"
)

var (
	Agent interfaces.Agent
	once  sync.Once
)

type agent struct {
	repo     *repository.Repositories
	Telegram interfaces.Telegram
}

func NewAgent() interfaces.Agent {
	once.Do(func() {
		Agent = &agent{
			repo: resource.GetRepositories(),
		}
	})
	return Agent
}

func (a *agent) TeleGramChat(modelName string, message []llms.MessageContent, user *domain.User) string {
	// agent
	agent := client.NewClient(manager.NewManager(user))
	res := agent.Chat(modelName, message)
	return res
}

func (a *agent) Chat(ctx context.Context, messages []domain.Message) (*domain.Message, error) {
	// 转换domain.Message为langchaingo的MessageContent
	var llmMessages []llms.MessageContent
	for _, msg := range messages {
		var msgType llms.ChatMessageType
		switch msg.Role {
		case "system":
			msgType = llms.ChatMessageTypeSystem
		case "user":
			msgType = llms.ChatMessageTypeHuman
		case "assistant":
			msgType = llms.ChatMessageTypeAI
		default:
			msgType = llms.ChatMessageTypeHuman
		}
		llmMessages = append(llmMessages, llms.TextParts(msgType, msg.Content))
	}

	// 从上下文中获取用户信息
	user := utils.GetUserFromContextOrDefault(ctx)

	// 获取默认agent配置
	agentConfig, err := a.repo.AgentConfig.GetAgentConfigByName("default")
	if err != nil {
		return nil, err
	}

	// 使用现有client进行聊天
	agentClient := client.NewClient(manager.NewManager(user))
	result := agentClient.Chat(agentConfig.Name, llmMessages)

	return &domain.Message{
		Role:    "assistant",
		Content: result,
	}, nil
}
