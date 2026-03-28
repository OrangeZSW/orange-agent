package file

import (
	"orange-agent/common"
)

var FileTools = []common.BaseTool{
	FileListTool,
	ReadFileTool,
	WriteFileTool,
	DeleteFileTool,
	RenameFileTool,
	CopyFileTool,
	FileSearchTool,
}
