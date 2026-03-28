package common

import "github.com/tmc/langchaingo/tools"

type BaseTool interface {
	tools.Tool
	Parameters() interface{}
}
