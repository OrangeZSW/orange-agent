package interfaces

import (
	"context"
	"orange-agent/domain"

	"github.com/tmc/langchaingo/llms"
)

// Telegram 定义Telegram机器人的核心接口
type Telegram interface {
	// SendMessage 发送消息给指定用户
	SendMessage(telegramId int64, text string)
	// Init 初始化Telegram机器人
	Init(config *domain.Telegram, handler MessageHandler) Client
}

// Manager 定义用户和消息管理接口
type Manager interface {
	// GetUser 获取或创建用户
	GetUser(telegramId int64, name string) *domain.User
	// GetMessageHistory 获取用户的消息历史
	GetMessageHistory(userId uint, limit int) []llms.MessageContent
}

// Client 定义Telegram客户端接口
type Client interface {
	// Init 初始化客户端配置
	Init(config *domain.Telegram)
	// Start 启动机器人
	Start()
	// Stop 停止机器人
	Stop()
	// SendMessage 发送消息
	SendMessage(telegramId int64, text string)
}

// MessageHandler 定义消息处理接口
type MessageHandler interface {
	// Handle 处理消息并返回响应
	Handle(ctx context.Context, modelName string, messages []llms.MessageContent) string
}
