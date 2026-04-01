// package task

// import (
// 	"context"
// 	"fmt"
// 	"orange-agent/domain"
// )

// // ExampleUsage 示例：如何使用任务编排系统
// func ExampleUsage() {
// 	// 2. 创建任务编排器
// 	config := DefaultOrchestratorConfig()
// 	config.WorkerCount = 3 // 设置3个worker并发执行
// 	// orchestrator := NewTaskOrchestrator(config)

// 	// 3. 创建总任务
// 	task := &domain.Task{
// 		SessionID:   "example-session-001",
// 		Description: "开发一个简单的待办事项应用，需要包含以下功能：\n1. 用户可以添加待办事项\n2. 用户可以标记待办事项为完成\n3. 用户可以删除待办事项\n4. 提供简单的Web界面",
// 		Status:      domain.StatusPending,
// 	}

// 	// 4. 执行任务
// 	ctx := context.Background()
// 	// result, err := orchestrator.Execute(ctx, task)
// 	if err != nil {
// 		fmt.Printf("任务执行失败: %v\n", err)
// 		return
// 	}

// 	// 5. 输出结果
// 	fmt.Println("任务执行完成！")
// 	fmt.Println("最终结果：")
// 	fmt.Println(result)
// }

// // ExampleWithCustomConfig 示例：使用自定义配置
// func ExampleWithCustomConfig() {
// 	// 创建自定义配置
// 	config := &OrchestratorConfig{
// 		WorkerCount:     5,  // 5个worker
// 		QueueBufferSize: 50, // 队列缓冲区大小50
// 	}

// 	// 创建编排器
// 	agentManager := NewSimpleAgentManager()
// 	orchestrator := NewTaskOrchestrator(agentManager, config)

// 	// 使用编排器...
// 	_ = orchestrator
// }
