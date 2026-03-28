package git

import "orange-agent/common"

var GitTools = []common.BaseTool{
	GitPushTool,
	GitDiffTool,
	GitCommitTool,
}
