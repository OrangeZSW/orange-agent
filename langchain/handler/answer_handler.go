package handler

import (
	"context"
	"fmt"

	"orange-agent/domain"
	"orange-agent/langchain/chain"
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

func (h *AnswerHandler) AnswerQuestion(user *domain.User, memory *domain.Memory, prompt string) string {
	ctx := context.Background()

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
