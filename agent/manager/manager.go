package manager

import (
	"orange-agent/agent/interfaces"
	"orange-agent/domain"
	"orange-agent/repository"
	"orange-agent/repository/resource"
	"orange-agent/telegram"
	"orange-agent/utils/file"
	"orange-agent/utils/logger"
	"strings"
	"sync"

	"github.com/tmc/langchaingo/llms"
	"go.yaml.in/yaml/v3"
)

var (
	skills []Skill
	once   sync.Once
)

type Skill struct {
	Name        string
	Description string
	Content     string
}

type manager struct {
	User     *domain.User
	log      *logger.Logger
	repo     *repository.Repositories
	telegram interfaces.Telegram
}

func NewManager(user *domain.User) interfaces.Manager {
	return &manager{
		log:      logger.GetLogger(),
		repo:     resource.GetRepositories(),
		User:     user,
		telegram: telegram.GetTelegram(),
	}
}

func (r *manager) SaveCallRecord(message []llms.MessageContent, resp *llms.ContentResponse, agentConfig *domain.AgentConfig) error {
	memory, _ := r.repo.Memory.GetMemoryByUserIdAndLimit(r.User.ID, 1)

	//获取当前用户问题
	callRecord := &domain.CallRecord{
		MemoryId:         memory[0].ID,
		UserID:           r.User.ID,
		AgentId:          agentConfig.ID,
		ModelName:        r.User.ModelName,
		PromptTokens:     resp.Choices[0].GenerationInfo["PromptTokens"].(int),
		CompletionTokens: resp.Choices[0].GenerationInfo["CompletionTokens"].(int),
		TotalTokens:      resp.Choices[0].GenerationInfo["TotalTokens"].(int),
	}
	return r.repo.AgentCallRecord.CreateAgentCallRecord(callRecord)
}

func (r *manager) TeleGramSendMessage(text string) {
	r.telegram.SendTeleGramMessage(int64(r.User.TelegramId), text)
}

func NewSkills() []Skill {
	once.Do(func() {
		files, err := file.GetFileList("../../")
		if err != nil {
			panic(err)
		}
		for _, item := range files {
			var skill Skill
			if item.Name == "SKILL.md" {
				content, _ := file.ReadFile(item.Path)
				parts := strings.SplitN(string(content), "---\n", 3)
				if len(parts) < 3 {
					continue
				}
				if err := yaml.Unmarshal([]byte(parts[1]), &skill); err != nil {
					continue
				}
				skill.Content = parts[2]
				skills = append(skills, skill)
			}
		}
	})
	return skills
}

func (r *manager) SystemPrompt() []llms.MessageContent {
	system := ""
	//当前系统架构
	agent, err := file.ReadFile("./AGENT.md")
	if err != nil {
		r.log.Error("获取系统AGENT.md失败: %v", err)
	}
	system = string(agent)
	message := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, system),
	}
	return message
}
