package agent

import (
	"orange-agent/common"
)

var AgentTools = []common.BaseTool{
	AgentAddTool,
	AgentRemoveTool,
	AgentListTool,
	AgentUpdateTool,
	AgentTestTool,
}
