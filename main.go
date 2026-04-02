package main

import (
	"orange-agent/agent"
	"orange-agent/agent/rag"
	"orange-agent/config"
	"orange-agent/repository/db"
	"orange-agent/repository/resource"
	"orange-agent/telegram"
	"orange-agent/utils/logger"
)

func main() {
	cfg := config.NewConfig()
	logger.InitDefaultLogger(cfg.Logger)
	resource.GetDataResource().Add(db.InitMysql(&cfg.Database), "mysql")

	// 初始化Redis向量存储
	ragConfig := &rag.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}
	if err := rag.InitializeWithRedis(ragConfig); err != nil {
		logger.GetLogger().Warn("Redis向量存储初始化失败: %v，代码搜索功能将不可用", err)
	}

	// 初始化Agent
	agentInstance := agent.NewAgent()

	// 初始化并启动Telegram
	tg := telegram.NewTelegram()
	client := tg.Init(&cfg.Telegram, agentInstance)
	client.Start()
}
