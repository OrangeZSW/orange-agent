package main

import (
	"context"
	"fmt"
	"log"
	"orange-agent/config"
	"orange-agent/langchain/llm"
	"orange-agent/repository/resource"
	"orange-agent/task/analyzer"
	task_context "orange-agent/task/context"
	"orange-agent/task/executor"
	"orange-agent/task/orchestrator"
	"orange-agent/task/summarizer"
	"orange-agent/utils/logger"
)

func main() {
	config := config.NewConfig()
	logger.InitDefaultLogger(config.Logger)
	resource.GetDataResource().InitMysql(&config.Database)

	// bot := telegram.NewTelegramBot(&config.Telegram)

	// bot.Start()
	TaskTest()

}

func TaskTest() {
	llm := llm.NewOpenAIProvider()
	llm.GetLLM("qwen3.5-27b")

	// 上下文管理器，每个子任务最大4000 tokens
	contextManager := task_context.NewContextManager(4000)

	// 分析器
	taskAnalyzer := analyzer.NewTaskAnalyzer(llm)

	// 执行器（3个并发worker）
	taskExecutor := executor.NewTaskExecutor(llm, contextManager, 3)

	// 总结器
	taskSummarizer := summarizer.NewTaskSummarizer(llm)

	// 编排器
	orchestrator := orchestrator.NewTaskOrchestrator(
		taskAnalyzer,
		taskExecutor,
		taskSummarizer,
		contextManager,
	)

	// 处理任务
	taskDescription := "帮我分析2024年AI行业趋势，并生成一份详细的报告，包括关键技术、市场应用和未来预测"

	ctx := context.Background()
	task, err := orchestrator.ProcessTask(ctx, "session_123", taskDescription)

	if err != nil {
		log.Fatalf("Task failed: %v", err)
	}

	fmt.Printf("Task Status: %s\n", task.Status)
	fmt.Printf("Result:\n%s\n", task.Result)
}
