package telegram

import (
	"orange-agent/domain"
	"orange-agent/repository/factory"
	"orange-agent/utils"
	"strings"

	tele "gopkg.in/telebot.v3"
)

type HandlerCommand struct {
	repoFactory *factory.Factory
	TelegramBot *TelegramBot
}

func NewHandlerCommand(bot *TelegramBot) *HandlerCommand {
	res := &HandlerCommand{
		TelegramBot: bot,
		repoFactory: factory.NewFactory(),
	}
	res.RegisterHandler()
	return res
}

//register handler

func (h *HandlerCommand) RegisterHandler() {
	h.TelegramBot.Bot.Handle("/start", h.Start)
	h.TelegramBot.Bot.Handle("/help", h.Help)
	h.TelegramBot.Bot.Handle("/addAgent", h.AddAgent)
	h.TelegramBot.Bot.Handle("/addModel", h.AddModel)
	h.TelegramBot.Bot.Handle("/agents", h.Agents)
	h.TelegramBot.Bot.Handle("/switch", h.Switch)
	h.TelegramBot.Bot.Handle("/model", h.Model)
}

func (tb *HandlerCommand) Start(c tele.Context) error {
	welcomeMsg := `欢迎使用 Orange Agent Bot！`
	return c.Send(welcomeMsg)
}

func (tb *HandlerCommand) Help(c tele.Context) error {
	helpMsg := `使用说明：
	1. 直接发送消息即可与 AI 对话
	2. /start - 显示欢迎信息
	3. /help - 显示帮助
	4. /agents - 显示所有代理信息
	5. /switch <agent_id> <model_index> - 切换代理
	6. /addAgent <agent_name> <base_url> <token> - 添加一个代理
	7. /addModel <agent_id> <model_name> - 添加一个模型
	8. /model - 显示当前使用的模型
	`
	return c.Send(helpMsg)
}

func (tb *HandlerCommand) AddAgent(c tele.Context) error {
	parts := strings.Fields(c.Message().Text)
	if len(parts) != 4 {
		return c.Reply("请输入正确的命令格式：/addAgent <agent_name> <base_url> <token>")
	}
	agentConfig := &domain.AgentConfig{
		Name:    parts[1],
		BaseUrl: parts[2],
		Token:   parts[3],
	}
	tb.repoFactory.AgentConfigRepo.CreateAgentConfig(agentConfig)
	return c.Reply("添加成功")
}

func (tb *HandlerCommand) Switch(c tele.Context) error {
	parts := strings.Fields(c.Message().Text)
	if len(parts) != 3 {
		return c.Reply("请输入正确的命令格式：/switch <agent_id> <model_index>")
	}
	agentConfig, _ := tb.repoFactory.AgentConfigRepo.GetAgentConfigById(utils.StrToUint(parts[1]))
	tb.repoFactory.UserRepo.UpdateUserModelName(c.Sender().ID, agentConfig.Models[utils.StrToUint(parts[2])-1])
	return c.Reply("切换成功")
}

func (tb *HandlerCommand) AddModel(c tele.Context) error {
	parts := strings.Fields(c.Message().Text)
	if len(parts) != 3 {
		return c.Reply("请输入正确的命令格式：/addModel <agent_id> <model_name>")
	}
	agentConfig, _ := tb.repoFactory.AgentConfigRepo.GetAgentConfigById(utils.StrToUint(parts[1]))
	agentConfig.Models = append(agentConfig.Models, parts[2])
	tb.repoFactory.AgentConfigRepo.UpdateAgentConfig(agentConfig)
	return c.Reply("添加成功")
}

func (tb *HandlerCommand) Agents(c tele.Context) error {
	agents, _ := tb.repoFactory.AgentConfigRepo.GetAllAgentConfig()
	var list string
	for _, agent := range agents {
		list += "id: " + utils.UintToStr(agent.ID) + " name: " + agent.Name + " models: " + strings.Join(agent.Models, ", ") + "\n"
	}
	return c.Reply("当前可用的代理：\n" + list)
}

func (tb *HandlerCommand) Model(c tele.Context) error {
	user, _ := tb.repoFactory.UserRepo.GetUserByTelegramId(c.Sender().ID)
	return c.Reply("当前使用的模型：" + user.ModelName)
}
