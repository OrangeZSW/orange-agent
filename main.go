package main

import (
	"orange-agent/agent"
	"orange-agent/agent/manager"
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

	// 初始化RAG模块
	if err := rag.Init(&rag.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}); err != nil {
		logger.GetLogger().Warn("RAG模块初始化失败: %v", err)
	}

	// 初始化并启动Telegram
	tg := telegram.NewTelegram()
	client := tg.Init(&cfg.Telegram, agent.NewAgent())

	// 设置Telegram实例到Agent管理器
	manager.SetTelegram(tg)

	client.Start()
}
