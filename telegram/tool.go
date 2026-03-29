package telegram

import (
	tele "gopkg.in/telebot.v3"
)

// SendMessage 主动给指定用户发送消息
func SendMessage(TeleGramId int64, text string) error {
	telegram := GetTelegramBot()

	// telebot.v3 的 Send 方法需要 Recipient 和要发送的内容
	recipient := &tele.User{ID: TeleGramId}

	// 发送消息
	_, err := telegram.Bot.Send(recipient, text)
	if err != nil {
		telegram.log.Error("主动发送消息失败, chatID: %d, error: %v", TeleGramId, err)
		return err
	}

	telegram.log.Info("主动发送消息成功, chatID: %d", TeleGramId)
	return nil
}
