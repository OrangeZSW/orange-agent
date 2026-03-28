package common

import (
	"context"
	"github.com/tmc/langchaingo/tools"
)

type BaseTool interface {
	tools.Tool
	Name() string
	Description() string
	Call(ctx context.Context, input string) (string, error)
	Parameters() interface{}
}
