package main

import (
	"orange-agent/config"
	dataresource "orange-agent/repository/data_resource"
	"orange-agent/telegram"
	"orange-agent/utils/logger"
)

func main() {
	config := config.NewConfig()
	logger.InitDefaultLogger(config.Logger)
	dataresource.GetDataResource().InitMysql(&config.Database)

	bot := telegram.NewTelegramBot(&config.Telegram)

	bot.Start()
}
