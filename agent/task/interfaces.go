package task

import (
	"context"
	"orange-agent/domain"
)

type TaskChat interface {
	TaskChat(ctx context.Context, messages []domain.Message) string
}
