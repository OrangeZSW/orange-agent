package telegram

import (
	"orange-agent/domain"
	"orange-agent/lanchain"
	"orange-agent/mysql"
	"orange-agent/utils"
	"orange-agent/utils/logger"

	tele "gopkg.in/telebot.v3"
)

type HandlerText struct {
	telegram *TelegramBot
	log      logger.Logger
	answer   *lanchain.AnswerHandler
	userSql  *mysql.UserSql
	lanchain *lanchain.Lnachain
}

func NewHandlerText(bot *TelegramBot) *HandlerText {
	res := &HandlerText{
		telegram: bot,
		log:      *logger.GetLogger(),
		answer:   lanchain.NewAnswerHandler(),
		userSql:  mysql.NewUserSql(),
		lanchain: lanchain.NewLnachain(),
	}
	res.RegisterHandler()
	return res
}

func (h *HandlerText) RegisterHandler() {
	h.telegram.Bot.Handle(tele.OnText, h.OnText)
}

func (h *HandlerText) OnText(c tele.Context) error {
	telegramId := c.Sender().ID
	username := c.Sender().Username
	user := h.GetUser(utils.Int64ToUint(telegramId), username)

	h.log.Info("收到用户 %d 输入: %s", telegramId, c.Text())
	res := h.answer.AnswerQuestion(*user, c.Text(), h.telegram.Config.Promete)
	h.log.Info("模型:%s 响应: %s", user.ModelName, res)
	return c.Reply(res)
}

func (h *HandlerText) GetUser(telegramId uint, username string) *domain.User {
	user, err := h.userSql.GetUserByTelegramId(int64(telegramId))
	if err != nil {
		h.log.Error("获取用户失败: %v", err)
		return nil
	}
	if user == nil {
		h.log.Error("用户不存在,创建用户")
		user = &domain.User{
			TelegramId: telegramId,
			Name:       username,
			ModelName:  h.lanchain.GetDefaultModelName(),
		}
		h.userSql.CreateUser(user)
		return user
	}
	return user
}
