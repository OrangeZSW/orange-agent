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
	"orange-agent/utils/logger"
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
	log      *logger.Logger
}

func NewAgent() interfaces.Agent {
	once.Do(func() {
		Agent = &agent{
			repo: resource.GetRepositories(),
			log:  logger.GetLogger(),
		}
	})
	return Agent
}

func (a *agent) TeleGramChat(ctx context.Context, modelName string, message []llms.MessageContent) string {
	// agent
	user, ok := utils.GetUserFromContext(ctx)
	if user == nil || !ok {
		a.log.Error("get user from context error")
		return "get user from context error"
	}
	client := client.NewClient(manager.NewManager(user))
	res := ""
	switch user.ChainMode {
	case domain.NORMAL:
		res = client.Chat(modelName, message)
	case domain.TASK:
		res = TaskChat()
	default:
		res = client.Chat(modelName, message)
	}
	return res
}

func (a *agent) Chat(ctx context.Context, messages []domain.Message) string {
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
	user, falg := utils.GetUserFromContext(ctx)
	if !falg && user == nil {
		a.log.Error("get user from context error")
		return "get user from context error"
	}

	// 使用现有client进行聊天
	agentClient := client.NewClient(manager.NewManager(user))
	result := agentClient.Chat(user.ModelName, llmMessages)

	return result
}

func TaskChat() string {
	return ""
}
