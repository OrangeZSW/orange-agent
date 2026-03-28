package common

import (
	"context"
)

type BaseTool struct {
	Name        string
	Description string
	Parameters  interface{}
	Call        func(ctx context.Context, input string) (string, error)
}
