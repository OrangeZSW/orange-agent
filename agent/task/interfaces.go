package task

import (
	"context"
	"orange-agent/domain"
)

type TaskChat interface {
	Chat(ctx context.Context, messages []domain.Message) string
}
