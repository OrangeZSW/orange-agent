package system

import "orange-agent/common"

var SystemTools = []common.BaseTool{
	&BuildTools{},
	&ProjectReboot{},
}
