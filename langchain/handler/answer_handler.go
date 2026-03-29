package handler

import (
	"context"
	"fmt"

	"orange-agent/domain"
	"orange-agent/langchain/chain"
	repo_factory "orange-agent/repository/factory"
	"orange-agent/utils"
	"orange-agent/utils/logger"

	"github.com/tmc/langchaingo/llms"
)

type AnswerHandler struct {
	chain      *chain.Chain
	repo       *repo_factory.Factory
	log        *logger.Logger
	currentMem *domain.Memory
}

func NewAnswerHandler() *AnswerHandler {
	return &AnswerHandler{
		chain: chain.NewChain(),
		repo:  repo_factory.NewFactory(),
		log:   logger.GetLogger(),
	}
}

func (h *AnswerHandler) AnswerQuestion(user *domain.User, memory *domain.Memory, prompt string) string {
	h.currentMem = memory
	ctx := context.Background()

	answer, err := h.chain.Process(ctx, user, memory.UserQuestion, prompt)
	if err != nil {
		h.log.Error("处理问题失败: %v", err)
		return fmt.Sprintf("系统错误: %v", err)
	}

	return answer
}

func (h *AnswerHandler) SaveCallRecord(user *domain.User, response *llms.ContentResponse, agentID uint, memoryID uint) error {
	if len(response.Choices) == 0 || response.Choices[0] == nil {
		h.log.Warn("无法保存空响应的调用记录")
		return nil
	}

	generationInfo := response.Choices[0].GenerationInfo
	callRecord := &domain.CallRecord{
		ModelName:        user.ModelName,
		AgentId:          agentID,
		UserID:           user.ID,
		CompletionTokens: utils.GetIntFromMap(generationInfo, "CompletionTokens"),
		PromptTokens:     utils.GetIntFromMap(generationInfo, "PromptTokens"),
		TotalTokens:      utils.GetIntFromMap(generationInfo, "TotalTokens"),
		MemoryId:         memoryID,
	}

	if err := h.repo.AgentCallRecordRepo.CreateAgentCallRecord(callRecord); err != nil {
		h.log.Error("保存调用记录失败: %v", err)
		return err
	}

	h.log.Info("调用记录已保存")
	return nil
}

func (h *AnswerHandler) GetDefaultModelName() string {
	return h.chain.GetDefaultModelName()
}
