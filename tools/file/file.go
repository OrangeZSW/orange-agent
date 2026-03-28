package file

import "orange-agent/common"

var FileTools = []common.BaseTool{
	&FileRead{},
	&FileList{},
	&FileWrite{},
	&FileDeleteTools{},
	&FileRenameTools{},
	&FileCopyTools{},
	&FileSearchTools{},
}
