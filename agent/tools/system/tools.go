package system

import (
	"orange-agent/common"
)

var SystemTools = []common.BaseTool{
	BuildTool,
	ProjectRebootTool,
	LogViewTool,
	EnvManageTool,
	DependencyCheckTool,
	PerformanceMonitorTool,
	ApiTesterTool,
	WebSearchTool,
	CodeSearchTool,
	CodeIndexInitTool,
}
