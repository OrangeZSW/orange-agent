package tools

import (
	"context"
	"time"
)

var (
	TimeTools = []BaseTool{
		&CurrTime{},
	}
)

type CurrTime struct {
	BaseTool
}

// Name
func (t *CurrTime) Name() string {
	return "curr_time"
}
func (t *CurrTime) Description() string {
	return "Get the current time,no input"
}
func (t *CurrTime) Call(ctx context.Context, input string) (string, error) {
	return time.Now().Format("2006-01-02 15:04:05"), nil
}

func (t *CurrTime) Parameters() interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{}, // 空参数
		"required":   []string{},               // 没有必需参数
	}
}
