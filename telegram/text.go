package telegram

import (
	"orange-agent/domain"
	"orange-agent/lanchain"
	"orange-agent/repository/factory"
	"orange-agent/utils"
	"orange-agent/utils/logger"

	tele "gopkg.in/telebot.v3"
)

type HandlerText struct {
	telegram    *TelegramBot
	log         logger.Logger
	answer      *lanchain.AnswerHandler
	lanchain    *lanchain.Lnachain
	repoFactory *factory.Factory
}

func NewHandlerText(bot *TelegramBot) *HandlerText {
	res := &HandlerText{
		telegram:    bot,
		log:         *logger.GetLogger(),
		answer:      lanchain.NewAnswerHandler(),
		lanchain:    lanchain.NewLnachain(),
		repoFactory: factory.NewFactory(),
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
	memory := &domain.Memory{
		UserId:       user.ID,
		UserQuestion: c.Text(),
	}
	h.repoFactory.MemoryRepo.CreateMemory(memory)
	h.log.Info("收到用户 %d 输入: %s", telegramId, c.Text())
	res := h.answer.AnswerQuestion(user, memory, h.telegram.Config.Promete)
	h.log.Info("模型:%s 响应: %s", user.ModelName, res)
	memory.AgentAnswer = res
	h.repoFactory.MemoryRepo.UpdateMemory(memory)
	return c.Reply(res)
}

func (h *HandlerText) GetUser(telegramId uint, username string) *domain.User {
	user, err := h.repoFactory.UserRepo.GetUserByTelegramId(int64(telegramId))
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
		h.repoFactory.UserRepo.CreateUser(user)
		return user
	}
	return user
}
