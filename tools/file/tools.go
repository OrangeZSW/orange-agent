package file

import (
	"orange-agent/common"
)

var FileTools = []common.BaseTool{
	FileReadTool,
	FileWriteTool,
	FileDeleteTool,
	FileListTool,
	FileRenameTool,
	FileSearchTool,
}
