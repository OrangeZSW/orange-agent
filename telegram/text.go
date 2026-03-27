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
	Bot      *tele.Bot
	log      logger.Logger
	answer   *lanchain.Answer
	userSql  *mysql.UserSql
	lanchain *lanchain.Lnachain
}

func NewHandlerText(bot *tele.Bot) *HandlerText {
	return &HandlerText{
		Bot:      bot,
		log:      *logger.GetLogger(),
		answer:   lanchain.NewAnswer(),
		userSql:  mysql.NewUserSql(),
		lanchain: lanchain.NewLnachain(),
	}
}

func (h *HandlerText) RegisterHandler() {
	h.Bot.Handle(tele.OnText, h.OnText)
}

func (h *HandlerText) OnText(c tele.Context) error {
	telegramId := c.Sender().ID
	username := c.Sender().Username
	user := h.GetUser(utils.Int64ToUint(telegramId), username)

	h.log.Info("收到用户 %d 输入: %s", telegramId, c.Text())
	res := h.answer.Answer(*user, c.Text(), "")
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
