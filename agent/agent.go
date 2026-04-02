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
	instance interfaces.Agent
	once     sync.Once
)

type agent struct {
	repo *repository.Repositories
	log  *logger.Logger
}

// NewAgent 创建Agent实例（单例模式）
func NewAgent() interfaces.Agent {
	once.Do(func() {
		instance = &agent{
			repo: resource.GetRepositories(),
			log:  logger.GetLogger(),
		}
	})
	return instance
}

// Handle 实现MessageHandler接口，处理Telegram消息
func (a *agent) Handle(ctx context.Context, modelName string, messages []llms.MessageContent) string {
	user, ok := utils.GetUserFromContext(ctx)
	if user == nil || !ok {
		a.log.Error("从上下文获取用户失败")
		return "获取用户信息失败"
	}

	c := client.NewClient(manager.NewManager(user))
	return c.Chat(modelName, messages)
}

// Chat 处理通用对话
func (a *agent) Chat(ctx context.Context, messages []domain.Message) string {
	// TODO: 实现通用对话处理
	return "功能开发中"
}
