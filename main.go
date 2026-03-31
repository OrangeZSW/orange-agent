package main

import (
	"orange-agent/agent"
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
	telegram.NewTelegram().InitTelegram(&config.Telegram, agent.NewAgent()).Start()
}
