package tools

import (
	"orange-agent/common"
	"orange-agent/tools/file"
	"orange-agent/tools/git"
	"orange-agent/tools/system"
	"sync"

	"github.com/tmc/langchaingo/llms"
)

var Tools []common.BaseTool

var Once sync.Once

func InitTools() {
	Once.Do(func() {
		Tools = append(Tools, file.FileTools...)
		Tools = append(Tools, TimeTools...)
		Tools = append(Tools, git.GitTools...)
		Tools = append(Tools, system.SystemTools...)
	})
}

func GetTools() map[string]common.BaseTool {
	InitTools()
	data := make(map[string]common.BaseTool, len(Tools))
	for _, tool := range Tools {
		data[tool.Name()] = tool
	}
	return data
}

func GetEllTools() []llms.Tool {
	InitTools()
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
