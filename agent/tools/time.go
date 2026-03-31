package tools

import (
	"context"
	"orange-agent/common"
	"time"
)

var CurrTimeTool = common.BaseTool{
	Name:        "curr_time",
	Description: "获取当前时间，无需输入参数",
	Parameters: map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	},
	Call: handlerCurrTime,
}

func handlerCurrTime(ctx context.Context, input string) (string, error) {
	return time.Now().Format("2006-01-02 15:04:05"), nil
}
