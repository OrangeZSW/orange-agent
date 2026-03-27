package tools

import (
	"sync"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
)

var Tools []BaseTool
var Once sync.Once

type BaseTool interface {
	tools.Tool
	Parameters() interface{}
}

func RegisterTools() {
	Once.Do(func() {
		Tools = append(Tools, TimeTools...)
	})
}

func GetTools() map[string]BaseTool {
	RegisterTools()
	data := make(map[string]BaseTool, len(Tools))
	for _, tool := range Tools {
		data[tool.Name()] = tool
	}
	return data
}

func GetEllTools() []llms.Tool {
	RegisterTools()
	llmTools := make([]llms.Tool, 0, len(Tools))
	for _, t := range Tools {
		// 为每个工具构建 llms.Tool 结构体
		llmTool := llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        t.Name(),        // 获取工具名称
				Description: t.Description(), // 获取工具描述
				Parameters:  t.Parameters(),
			},
		}
		llmTools = append(llmTools, llmTool)
	}

	return llmTools
}
