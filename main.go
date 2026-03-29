package main

import (
	"orange-agent/config"
	"orange-agent/repository/resource"
	"orange-agent/telegram"
	"orange-agent/utils/logger"
)

func main() {
	config := config.NewConfig()
	logger.InitDefaultLogger(config.Logger)
	resource.GetDataResource().InitMysql(&config.Database)

	bot := telegram.NewTelegramBot(&config.Telegram)

	bot.Start()
}
