package git

import "orange-agent/common"

var GitTools = []common.BaseTool{
	&GitDiff{},
	&GitCommit{},
	&GitPush{},
}
