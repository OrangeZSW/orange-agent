package database

import (
	"orange-agent/common"
)

var DatabaseTools = []common.BaseTool{
	DatabaseQueryTool,
	DatabaseExecuteTool,
}
