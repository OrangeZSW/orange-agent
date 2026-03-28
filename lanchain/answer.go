package lanchain

import (
	"context"
	"encoding/json"
	"orange-agent/domain"
	"orange-agent/mysql"
	"orange-agent/tools"
	"orange-agent/utils"
	"orange-agent/utils/logger"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type Answer struct {
	memorySql          *mysql.MemorySql
	lanchain           *Lnachain
	log                *logger.Logger
	agentCallRecordSql *mysql.AgentCallRecordSql
}

// New
func NewAnswer() *Answer {
	return &Answer{
		memorySql:          mysql.NewMemorySql(),
		lanchain:           NewLnachain(),
		log:                logger.GetLogger(),
		agentCallRecordSql: mysql.NewAgentCallRecordSql(),
	}
}

// 统一调用接口
func (l *Answer) Answer(user domain.User, question string, promete string) string {
	ctx := context.Background()
	llm := l.lanchain.GetLLM(user.ModelName)
	messages := l.buildMessages(user, question, promete)
	l.log.Info("开始调用模型[%s][%s]", l.lanchain.agentConfig.Name, user.ModelName)
	l.log.Info("当前系统工具:%v", tools.GetTools())
	answer, err := llm.GenerateContent(ctx, messages, llms.WithTools(tools.GetEllTools()))
	if err != nil {
		l.log.Error("调用模型失败: %v", err)
		return err.Error()
	}
	if len(answer.Choices) == 0 {
		return ""
	}
	choices := answer.Choices[0]

	if choices != nil && len(choices.ToolCalls) > 0 {
		return l.HandlerTools(ctx, user, messages, answer, llm)
	}

	l.saveCallRecord(user, answer)
	return choices.Content
}

// 调用工具
func (l *Answer) HandlerTools(ctx context.Context, user domain.User, message []llms.MessageContent, answer *llms.ContentResponse, llm *openai.LLM) string {
	ToolsCalls := answer.Choices[0].ToolCalls

	//添加assistant
	AiMessage := llms.MessageContent{
		Role:  llms.ChatMessageTypeAI,
		Parts: []llms.ContentPart{},
	}

	//出口
	if len(ToolsCalls) == 0 {
		AiMessage = llms.MessageContent{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextPart(answer.Choices[0].Content),
			},
		}
		return answer.Choices[0].Content
	}
	ToolsMessage := llms.MessageContent{
		Role:  llms.ChatMessageTypeTool,
		Parts: []llms.ContentPart{},
	}
	// 执行工具
	for _, toolrecall := range ToolsCalls {

		AiMessage.Parts = append(AiMessage.Parts, llms.ToolCall{
			ID:   toolrecall.ID,
			Type: toolrecall.Type,
			FunctionCall: &llms.FunctionCall{
				Name:      toolrecall.FunctionCall.Name,
				Arguments: toolrecall.FunctionCall.Arguments,
			},
		})

		l.log.Info("执行工具调用：%-10s,参数：%.20s", toolrecall.FunctionCall.Name, toolrecall.FunctionCall.Arguments)
		// 解析参数（假设参数是JSON字符串）
		var args map[string]interface{}
		if err := json.Unmarshal([]byte(toolrecall.FunctionCall.Arguments), &args); err != nil {
			l.log.Error("解析参数失败:%v", err)
			ToolsMessage.Parts = append(ToolsMessage.Parts, llms.ToolCallResponse{
				ToolCallID: toolrecall.ID,
				Content:    "解析参数失败" + err.Error(),
				Name:       toolrecall.FunctionCall.Name,
			})
			continue
		}

		res, err := l.executeTool(ctx, toolrecall.FunctionCall.Name, toolrecall.FunctionCall.Arguments)
		if err != nil {
			l.log.Error("执行工具失败: %v", err)
			ToolsMessage.Parts = append(ToolsMessage.Parts, llms.ToolCallResponse{
				ToolCallID: toolrecall.ID,
				Content:    "执行工具失败" + err.Error(),
				Name:       toolrecall.FunctionCall.Name,
			})
			return "执行工具失败" + err.Error()
		} else {
			l.log.Info("工具调用成功：%-10s,结果：%.20s", toolrecall.FunctionCall.Name, res)
			ToolsMessage.Parts = append(ToolsMessage.Parts, llms.ToolCallResponse{
				ToolCallID: toolrecall.ID,
				Content:    res,
				Name:       toolrecall.FunctionCall.Name,
			})
		}
	}

	message = append(message, AiMessage)
	message = append(message, ToolsMessage)

	l.log.Info("工具调用结束")
	l.log.Info("构建的消息: %.5v", message)

	//调用模型
	answer, err := llm.GenerateContent(ctx, message, llms.WithTools(tools.GetEllTools()))
	if err != nil {
		l.log.Error("调用模型失败: %v", err)
		return "调用模型失败"
	}
	l.saveCallRecord(user, answer)

	return l.HandlerTools(ctx, user, message, answer, llm)
}

// 执行工具
func (a *Answer) executeTool(ctx context.Context, name string, input string) (string, error) {
	data := tools.GetTools()
	for _, tool := range tools.GetEllTools() {
		if tool.Function.Name == name {
			res, err := data[name].Call(ctx, input)
			if err != nil {
				return "", err
			} else {
				return res, nil
			}
		}
	}
	a.log.Error("未找到工具：%s", name)
	return name + "工具未找到", nil
}

// 构建消息
func (l *Answer) buildMessages(user domain.User, question string, promete string) []llms.MessageContent {
	var messages []llms.MessageContent

	messages = append(messages, llms.TextParts(llms.ChatMessageTypeSystem, promete))

	memory, err := l.memorySql.GetMemoryByUserId(user.ID)
	l.log.Debug("用户记忆：%v", *memory)
	if err != nil {
		l.log.Error("获取用户记忆失败: %v", err)
	}
	for _, m := range *memory {
		messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, m.UserQuestion))
		messages = append(messages, llms.TextParts(llms.ChatMessageTypeAI, m.AgentAnswer))
	}
	messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, question))

	l.log.Debug("构建的消息：%v", messages)
	return messages
}

// 构建 agent call record
func (l *Answer) saveCallRecord(user domain.User, answer *llms.ContentResponse) {
	originMap := answer.Choices[0].GenerationInfo
	agentCallRecord := &domain.CallRecord{
		AgentName:        user.ModelName,
		AgentId:          l.lanchain.agentConfig.ID,
		UserID:           user.ID,
		CompletionTokens: utils.GetIntFromMap(originMap, "CompletionTokens"),
		PromptTokens:     utils.GetIntFromMap(originMap, "PromptTokens"),
		TotalTokens:      utils.GetIntFromMap(originMap, "TotalTokens"),
	}
	l.agentCallRecordSql.CreateAgentCallRecord(agentCallRecord)
}
