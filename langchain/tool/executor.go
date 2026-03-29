package tool

import (
	"context"
	"fmt"

	"orange-agent/tools"
	"orange-agent/utils/logger"
)

type Executor struct {
	log *logger.Logger
}

func NewExecutor() *Executor {
	return &Executor{
		log: logger.GetLogger(),
	}
}

func (e *Executor) Execute(ctx context.Context, toolName string, arguments string) (string, error) {
	availableTools := tools.GetTools()

	if _, exists := availableTools[toolName]; !exists {
		e.log.Error("未找到工具：%s", toolName)
		return "", fmt.Errorf("工具 '%s' 不存在", toolName)
	}

	result, err := availableTools[toolName].Call(ctx, arguments)
	if err != nil {
		return "", fmt.Errorf("工具调用失败: %w", err)
	}

	e.log.Info("工具 %s 执行成功", toolName)
	return result, nil
}

func (e *Executor) GetAvailableTools() map[string]interface{} {
	toolsMap := tools.GetTools()
	result := make(map[string]interface{})
	for k, v := range toolsMap {
		result[k] = v
	}
	return result
}
