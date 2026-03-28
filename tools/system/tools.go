package system

import (
	"orange-agent/common"
)

var SystemTools = []common.BaseTool{
	BuildTool,
	ProjectRebootTool,
	LogViewTool,
	EnvManageTool,
	TestRunTool,
	DependencyCheckTool,
	PerformanceMonitorTool,
	ApiTesterTool,
}
