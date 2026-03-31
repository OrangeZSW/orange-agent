package agent

import (
	"orange-agent/agent/client"
	"orange-agent/agent/interfaces"
	"orange-agent/agent/manager"
	"orange-agent/domain"
	"orange-agent/repository"
	"orange-agent/repository/resource"

	"github.com/tmc/langchaingo/llms"
)

type agent struct {
	repo repository.Repositories
}

func NewAgent() interfaces.Agent {
	return &agent{
		repo: *resource.GetRepositories(),
	}
}

func (a *agent) TeleGramChat(modelNmae string, message []llms.MessageContent, user *domain.User) string {
	// memory
	question := ""
	testPart, ok := message[len(message)-1].Parts[0].(llms.TextContent)
	if ok {
		question = testPart.Text
	}
	memory := &domain.Memory{
		UserQuestion: question,
		UserId:       user.ID,
	}
	a.repo.Memory.CreateMemory(memory)

	// agent
	agent := client.NewClient(manager.NewManager(user, memory))
	res := agent.Chat(modelNmae, message)

	//update memory
	memory.AgentAnswer = res
	a.repo.Memory.UpdateMemory(memory)
	return res
}
