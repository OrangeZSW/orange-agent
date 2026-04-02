package system

import (
	"orange-agent/common"
)

var SystemTools = []common.BaseTool{
	BuildTool,
	ProjectRebootTool,
	EnvManageTool,
	CodeSearchTool,
	CodeIndexInitTool,
	CurrTimeTool,
}
