package lanchain

import (
	"context"
	"encoding/json"
	"fmt"
	"orange-agent/domain"
	"orange-agent/mysql"
	"orange-agent/tools"
	"orange-agent/utils"
	"orange-agent/utils/logger"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// AnswerHandler 处理用户问题的答案生成
type AnswerHandler struct {
	memorySql          *mysql.MemorySql
	langChain          *Lnachain
	logger             *logger.Logger
	agentCallRecordSql *mysql.AgentCallRecordSql
	menmory            *domain.Memory
}

func NewAnswerHandler() *AnswerHandler {
	return &AnswerHandler{
		memorySql:          mysql.NewMemorySql(),
		langChain:          NewLnachain(),
		logger:             logger.GetLogger(),
		agentCallRecordSql: mysql.NewAgentCallRecordSql(),
	}
}

// CallLLM 调用语言模型生成答案
func (h *AnswerHandler) CallLLM(ctx context.Context, messages []llms.MessageContent, llm *openai.LLM, user domain.User) (*llms.ContentResponse, error) {
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
func (h *AnswerHandler) AnswerQuestion(user domain.User, memory *domain.Memory, prompt string) string {
	h.menmory = memory
	ctx := context.Background()
	llm := h.langChain.GetLLM(user.ModelName)
	messages := h.buildMessages(user, memory.UserQuestion, prompt)

	h.logger.Info("准备调用模型[%s][%s]", h.langChain.agentConfig.Name, user.ModelName)
	h.logger.Info("可用工具列表: %v", tools.GetTools())

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

// handleToolCalls 递归处理工具调用
func (h *AnswerHandler) handleToolCalls(ctx context.Context, user domain.User, messages []llms.MessageContent,
	response *llms.ContentResponse, llm *openai.LLM) *llms.ContentResponse {

	choice := response.Choices[0]
	if choice == nil || len(choice.ToolCalls) == 0 {
		return response
	}

	// 构建包含工具调用的消息
	updatedMessages, err := h.buildToolMessages(ctx, choice.ToolCalls, response, messages)
	if err != nil {
		h.logger.Error("构建工具消息失败: %v", err)
		return h.createErrorResponse(response, fmt.Sprintf("构建工具消息失败: %v", err))
	}

	// 再次调用模型
	newResponse, err := h.CallLLM(ctx, updatedMessages, llm, user)
	if err != nil {
		h.logger.Error("工具调用后再次调用模型失败: %v", err)
		return h.createErrorResponse(response, fmt.Sprintf("工具调用失败: %v", err))
	}

	// 递归处理可能的进一步工具调用
	return h.handleToolCalls(ctx, user, updatedMessages, newResponse, llm)
}

// buildToolMessages 构建包含工具调用和响应的消息
func (h *AnswerHandler) buildToolMessages(ctx context.Context, toolCalls []llms.ToolCall,
	response *llms.ContentResponse, messages []llms.MessageContent) ([]llms.MessageContent, error) {

	aiMessage := llms.MessageContent{
		Role:  llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{},
	}

	toolMessage := llms.MessageContent{
		Role:  llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{},
	}

	// 执行每个工具调用
	for _, toolCall := range toolCalls {
		// 添加AI的工具调用消息
		aiMessage.Parts = append(aiMessage.Parts, llms.ToolCall{
			ID:   toolCall.ID,
			Type: toolCall.Type,
			FunctionCall: &llms.FunctionCall{
				Name:      toolCall.FunctionCall.Name,
				Arguments: toolCall.FunctionCall.Arguments,
			},
		})

		h.logger.Info("执行工具调用：%s，参数：%.200s",
			toolCall.FunctionCall.Name,
			toolCall.FunctionCall.Arguments)

		// 执行工具
		result, err := h.executeTool(ctx, toolCall.FunctionCall.Name, toolCall.FunctionCall.Arguments)
		if err != nil {
			h.logger.Error("执行工具 %s 失败: %v", toolCall.FunctionCall.Name, err)
			toolMessage.Parts = append(toolMessage.Parts, llms.ToolCallResponse{
				ToolCallID: toolCall.ID,
				Content:    fmt.Sprintf("工具执行失败: %v", err),
				Name:       toolCall.FunctionCall.Name,
			})
		} else {
			h.logger.Info("工具调用 %s 成功，结果：%.50s", toolCall.FunctionCall.Name, result)
			toolMessage.Parts = append(toolMessage.Parts, llms.ToolCallResponse{
				ToolCallID: toolCall.ID,
				Content:    result,
				Name:       toolCall.FunctionCall.Name,
			})
		}
	}

	h.logger.Info("工具调用处理完成")

	// 构建新的消息列表
	updatedMessages := make([]llms.MessageContent, 0, len(messages)+2)
	updatedMessages = append(updatedMessages, messages...)
	updatedMessages = append(updatedMessages, aiMessage, toolMessage)

	return updatedMessages, nil
}

// executeTool 执行具体的工具
func (h *AnswerHandler) executeTool(ctx context.Context, toolName string, arguments string) (string, error) {
	availableTools := tools.GetTools()

	// 检查工具是否存在
	if _, exists := availableTools[toolName]; !exists {
		h.logger.Error("未找到工具：%s", toolName)
		return "", fmt.Errorf("工具 '%s' 不存在", toolName)
	}

	// 执行工具
	result, err := availableTools[toolName].Call(ctx, arguments)
	if err != nil {
		return "", fmt.Errorf("工具调用失败: %w", err)
	}

	return result, nil
}

// buildMessages 构建完整的对话消息
func (h *AnswerHandler) buildMessages(user domain.User, question string, prompt string) []llms.MessageContent {
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, prompt),
	}

	// 添加用户记忆
	memories, err := h.memorySql.GetMemoryByUserId(user.ID)
	if err != nil {
		h.logger.Error("获取用户记忆失败: %v", err)
	} else {
		h.logger.Debug("加载用户记忆：%d 条", len(*memories))

		startIdx := 0
		if len(*memories) > 5 {
			startIdx = len(*memories) - 5
		}
		for _, memory := range (*memories)[startIdx:] {
			messages = append(messages,
				llms.TextParts(llms.ChatMessageTypeHuman, memory.UserQuestion),
				llms.TextParts(llms.ChatMessageTypeAI, memory.AgentAnswer),
			)
		}
	}

	// 添加当前问题
	messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, question))

	h.logger.Debug("构建的消息数量：%d", len(messages))
	return messages
}

// saveCallRecord 保存调用记录
func (h *AnswerHandler) saveCallRecord(user domain.User, response *llms.ContentResponse) {
	if len(response.Choices) == 0 || response.Choices[0] == nil {
		h.logger.Warn("无法保存空响应的调用记录")
		return
	}

	generationInfo := response.Choices[0].GenerationInfo
	callRecord := &domain.CallRecord{
		AgentName:        user.ModelName,
		AgentId:          h.langChain.agentConfig.ID,
		UserID:           user.ID,
		CompletionTokens: utils.GetIntFromMap(generationInfo, "CompletionTokens"),
		PromptTokens:     utils.GetIntFromMap(generationInfo, "PromptTokens"),
		TotalTokens:      utils.GetIntFromMap(generationInfo, "TotalTokens"),
		MenmoryId:        h.menmory.ID,
	}

	if err := h.agentCallRecordSql.CreateAgentCallRecord(callRecord); err != nil {
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

// ToolArguments 工具调用的参数结构
type ToolArguments struct {
	Input string `json:"input"`
}

// parseToolArguments 解析工具参数
func parseToolArguments(jsonArgs string) (ToolArguments, error) {
	var args ToolArguments
	if err := json.Unmarshal([]byte(jsonArgs), &args); err != nil {
		return args, fmt.Errorf("解析工具参数失败: %w", err)
	}
	return args, nil
}
