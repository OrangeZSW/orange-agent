package telegram

import "orange-agent/telegram"

func SendTelegramMessage(TeleGramId int64, text string) {
	telegram.SendMessage(TeleGramId, text)
}
