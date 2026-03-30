package handler

import (
	"context"
	"fmt"
	"time"

	"orange-agent/domain"
	"orange-agent/langchain/chain"
	"orange-agent/langchain/interfaces"
	"orange-agent/utils"
	"orange-agent/utils/logger"
)

type AnswerHandler struct {
	chain *chain.Chain
	log   *logger.Logger
}

func NewAnswerHandler() *AnswerHandler {
	return &AnswerHandler{
		chain: chain.NewChain(),
		log:   logger.GetLogger(),
	}
}

func (h *AnswerHandler) SetMessageSender(sender interfaces.MessageSender) {
	h.chain.SetMessageSender(sender)
}

func (h *AnswerHandler) AnswerQuestion(user *domain.User, memory *domain.Memory, prompt string) string {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	go h.watchTimeout(ctx, user)

	answer, err := h.chain.Process(ctx, user, memory.ID, memory.UserQuestion, prompt)
	if err != nil {
		h.log.Error("处理问题失败: %v", err)
		return fmt.Sprintf("系统错误: %v", err)
	}

	return answer
}

func (h *AnswerHandler) GetDefaultModelName() string {
	return h.chain.GetDefaultModelName()
}

// 超时提示
func (h *AnswerHandler) watchTimeout(ctx context.Context, user *domain.User) {
	select {
	case <-ctx.Done():
		// 只有在超时或取消时才执行
		if ctx.Err() == context.DeadlineExceeded {
			timeoutMsg := "模型响应超时，请稍后再试。"
			h.log.Warn("用户 %d 的请求超时", user.ID)
			if h.chain.MenangerSender != nil {
				h.chain.MenangerSender.SendMessage(utils.UintToInt64(user.TelegramId), timeoutMsg)
			}
		}
	case <-time.After(61 * time.Second):
		// 安全保护：防止 goroutine 泄漏
		h.log.Debug("超时监控 goroutine 正常退出")
	}
}
