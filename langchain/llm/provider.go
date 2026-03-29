package llm

import (
	"context"

	"github.com/tmc/langchaingo/llms"
)

type Provider interface {
	GetLLM(model string) (interface{}, error)
	GetDefaultModelName() string
	Call(ctx context.Context, messages []llms.MessageContent, tools []llms.Tool) (*llms.ContentResponse, error)
	GetCurrentConfig() interface{}
}
