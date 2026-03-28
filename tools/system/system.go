package system

import "orange-agent/common"

var (
	BuildTool              = &BuildTools{}
	ProjectRebootTool      = &ProjectReboot{}
	LogViewTool            = &LogViewTools{}
	EnvManageTool          = &EnvManageTools{}
	TestRunTool            = &TestRunTools{}
	DependencyCheckTool    = &DependencyCheckTools{}
	PerformanceMonitorTool = &PerformanceMonitorTools{}
	ApiTesterTool          = &ApiTesterTools{}
	ConfigValidatorTool    = &ConfigValidatorTools{}
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
	ConfigValidatorTool,
}
