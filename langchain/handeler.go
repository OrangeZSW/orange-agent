package langchain

import (
	"context"
	"fmt"
	"orange-agent/domain"
	repo_factory "orange-agent/repository/factory"
	"orange-agent/tools"
	"orange-agent/utils"
	"orange-agent/utils/logger"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// AnswerHandler 处理用户问题的答案生成
type AnswerHandler struct {
	langChain *Lnachain
	logger    *logger.Logger
	memory    *domain.Memory
	repo      *repo_factory.Factory
}

func NewAnswerHandler() *AnswerHandler {
	return &AnswerHandler{
		langChain: NewLnachain(),
		logger:    logger.GetLogger(),
		repo:      repo_factory.NewFactory(),
	}
}

// CallLLM 调用语言模型生成答案
func (h *AnswerHandler) CallLLM(ctx context.Context, messages []llms.MessageContent, llm *openai.LLM, user *domain.User) (*llms.ContentResponse, error) {
	response, err := llm.GenerateContent(ctx, messages, llms.WithTools(tools.GetEllTools()))
	if err != nil {
		h.logger.Error("调用语言模型失败: %v", err)
		return nil, fmt.Errorf("调用语言模型失败: %w", err)
	}

	h.saveCallRecord(user, response)
	h.logger.Info("语言模型调用成功，已保存调用记录")
	return response, nil
}

// AnswerQuestion 处理用户问题并返回答案
func (h *AnswerHandler) AnswerQuestion(user *domain.User, memory *domain.Memory, prompt string) string {
	h.memory = memory
	ctx := context.Background()
	llm := h.langChain.GetLLM(user.ModelName)
	messages := h.buildMessages(user, memory.UserQuestion, prompt)

	h.logger.Info("准备调用模型[%s][%s]", h.langChain.agentConfig.Name, user.ModelName)

	response, err := h.CallLLM(ctx, messages, llm, user)
	if err != nil {
		h.logger.Error("调用模型失败: %v", err)
		return fmt.Sprintf("系统错误: %v", err)
	}

	if len(response.Choices) == 0 {
		h.logger.Warn("模型返回空选择列表")
		return "抱歉，我没有收到有效的回复"
	}

	choice := response.Choices[0]
	if choice == nil {
		h.logger.Warn("模型返回空选择")
		return "抱歉，我没有收到有效的回复"
	}

	// 处理工具调用
	if len(choice.ToolCalls) > 0 {
		response = h.handleToolCalls(ctx, user, messages, response, llm)
	}

	if len(response.Choices) == 0 || response.Choices[0] == nil {
		return "抱歉，处理过程中出现错误"
	}

	return response.Choices[0].Content
}

// saveCallRecord 保存调用记录
func (h *AnswerHandler) saveCallRecord(user *domain.User, response *llms.ContentResponse) {
	if len(response.Choices) == 0 || response.Choices[0] == nil {
		h.logger.Warn("无法保存空响应的调用记录")
		return
	}

	generationInfo := response.Choices[0].GenerationInfo
	callRecord := &domain.CallRecord{
		ModelName:        user.ModelName,
		AgentId:          h.langChain.agentConfig.ID,
		UserID:           user.ID,
		CompletionTokens: utils.GetIntFromMap(generationInfo, "CompletionTokens"),
		PromptTokens:     utils.GetIntFromMap(generationInfo, "PromptTokens"),
		TotalTokens:      utils.GetIntFromMap(generationInfo, "TotalTokens"),
		MemoryId:         h.memory.ID,
	}

	if err := h.repo.AgentCallRecordRepo.CreateAgentCallRecord(callRecord); err != nil {
		h.logger.Error("保存调用记录失败: %v", err)
	}
}

// createErrorResponse 创建错误响应
func (h *AnswerHandler) createErrorResponse(originalResponse *llms.ContentResponse, errorMessage string) *llms.ContentResponse {
	if len(originalResponse.Choices) == 0 {
		originalResponse.Choices = []*llms.ContentChoice{
			{
				Content: errorMessage,
			},
		}
	} else if originalResponse.Choices[0] != nil {
		originalResponse.Choices[0].Content = errorMessage
	}

	return originalResponse
}
