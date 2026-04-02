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
	config := config.NewConfig()
	logger.InitDefaultLogger(config.Logger)
	resource.GetDataResource().Add(db.InitMysql(&config.Database), "mysql")

	// 初始化Redis向量存储
	ragConfig := &rag.RedisConfig{
		Host:     config.Redis.Host,
		Port:     config.Redis.Port,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	}
	if err := rag.InitializeWithRedis(ragConfig); err != nil {
		logger.GetLogger().Warn("Redis向量存储初始化失败: %v，代码搜索功能将不可用", err)
	}

	telegram := telegram.NewTelegram()
	telegram.InitTelegram(&config.Telegram, agent.NewAgent()).Start()
}
