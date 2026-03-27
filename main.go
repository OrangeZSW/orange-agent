package main

import (
	"orange-agent/config"
	"orange-agent/mysql"
	"orange-agent/telegram"
	"orange-agent/utils/logger"
)

func main() {
	config := config.NewConfig()
	logger.InitDefaultLogger(config.Logger)
	mysql.NewMysql(&config.Database)

	bot := telegram.NewTelegramBot(&config.Telegram)

	bot.Start()
}
