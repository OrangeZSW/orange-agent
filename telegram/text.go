package telegram

import (
	"orange-agent/domain"
	"orange-agent/langchain/chain"
	"orange-agent/langchain/handler"
	factory "orange-agent/repository/factory"
	"orange-agent/utils"
	"orange-agent/utils/logger"
	"strconv"

	tele "gopkg.in/telebot.v3"
)

type HandlerText struct {
	telegram *TelegramBot
	log      logger.Logger
	answer   *handler.AnswerHandler
	chain    *chain.Chain
	repo     *factory.Factory
}

func NewHandlerText(bot *TelegramBot) *HandlerText {
	res := &HandlerText{
		telegram: bot,
		log:      *logger.GetLogger(),
		answer:   handler.NewAnswerHandler(),
		chain:    chain.NewChain(),
		repo:     factory.NewFactory(),
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
	h.repo.MemoryRepo.CreateMemory(memory)
	h.log.Info("收到用户 %d 输入: %s", telegramId, c.Text())
	res := h.answer.AnswerQuestion(user, memory, h.telegram.Config.Promete)
	h.log.Info("模型:%s 响应: %s", user.ModelName, res)
	memory.AgentAnswer = res
	h.repo.MemoryRepo.UpdateMemory(memory)

	callRecord, err := h.repo.AgentCallRecordRepo.SelectByMemoryId(memory.ID)
	if err != nil {
		h.log.Error("获取调用记录失败: %v", err)
	}
	totalTokens := 0
	for _, record := range callRecord {
		totalTokens += record.TotalTokens
	}

	res = res + "\n\n使用toknen数：" + strconv.Itoa(totalTokens)
	return c.Reply(res, tele.ModeHTML)
}

func (h *HandlerText) GetUser(telegramId uint, username string) *domain.User {
	user, err := h.repo.UserRepo.GetUserByTelegramId(int64(telegramId))
	if err != nil {
		h.log.Error("获取用户失败: %v", err)
		return nil
	}
	if user == nil {
		h.log.Error("用户不存在,创建用户")
		user = &domain.User{
			TelegramId: telegramId,
			Name:       username,
			ModelName:  h.chain.GetDefaultModelName(),
		}
		h.repo.UserRepo.CreateUser(user)
		return user
	}
	return user
}
