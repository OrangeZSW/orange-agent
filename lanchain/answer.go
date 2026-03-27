package lanchain

import (
	"context"
	"orange-agent/domain"
	"orange-agent/mysql"
	"orange-agent/utils"
	"orange-agent/utils/logger"

	"github.com/tmc/langchaingo/llms"
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
	answer, err := llm.GenerateContent(ctx, messages)
	if err != nil {
		l.log.Error("调用模型失败: %v", err)
		return err.Error()
	}
	l.saveCallRecord(user, answer)
	return answer.Choices[0].Content
}

// 构建消息
func (l *Answer) buildMessages(user domain.User, question string, promete string) []llms.MessageContent {
	var messages []llms.MessageContent

	messages = append(messages, llms.TextParts(llms.ChatMessageTypeSystem, promete))

	memory, err := l.memorySql.GetMemoryByUserId(user.ID)
	logger.Debug("用户记忆：%v", *memory)
	if err != nil {
		logger.Error("获取用户记忆失败: %v", err)
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
		AgentName:        l.lanchain.GetDefaultModelName(),
		AgentId:          l.lanchain.agentConfig.ID,
		UserID:           user.ID,
		CompletionTokens: utils.GetIntFromMap(originMap, "CompletionTokens"),
		PromptTokens:     utils.GetIntFromMap(originMap, "PromptTokens"),
		TotalTokens:      utils.GetIntFromMap(originMap, "TotalTokens"),
	}
	l.agentCallRecordSql.CreateAgentCallRecord(agentCallRecord)
}
