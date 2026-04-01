package task

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"orange-agent/domain"
	"orange-agent/utils/logger"
)

// DAGEngine 依赖图执行引擎
type DAGEngine struct {
	taskChat       TaskChat
	contextManager *ContextManager
	maxWorkers     int
}

// NewDAGEngine 创建新的DAG引擎
func NewDAGEngine(taskChat TaskChat, maxWorkers int) *DAGEngine {
	return &DAGEngine{
		taskChat:       taskChat,
		contextManager: NewContextManager(),
		maxWorkers:     maxWorkers,
	}
}

// ExecuteDAG 执行有向无环图任务
func (de *DAGEngine) ExecuteDAG(ctx context.Context, task *domain.Task) (string, error) {
	logger.Info("开始执行DAG任务: %s", task.Description)

	// 1. 构建DAG图
	dag, err := de.buildDAG(task.Subtasks)
	if err != nil {
		return "", fmt.Errorf("构建DAG失败: %w", err)
	}

	// 2. 拓扑排序
	topology, err := de.topologicalSort(dag)
	if err != nil {
		return "", fmt.Errorf("拓扑排序失败: %w", err)
	}

	// 3. 按拓扑顺序执行
	result, err := de.executeTopology(ctx, dag, topology, task)
	if err != nil {
		return "", fmt.Errorf("执行任务失败: %w", err)
	}

	// 4. 更新任务状态
	task.Status = domain.StatusCompleted
	task.Result = result

	logger.Info("DAG任务执行完成")
	return result, nil
}

// buildDAG 构建有向无环图
func (de *DAGEngine) buildDAG(subTasks []*domain.SubTask) (*domain.DependencyGraph, error) {
	dag := &domain.DependencyGraph{
		Nodes:    make([]*domain.DAGNode, 0, len(subTasks)),
		Edges:    make([]*domain.DAGEdge, 0),
		Metadata: make(map[string]any),
	}

	// 创建节点映射
	nodeMap := make(map[string]*domain.DAGNode)
	for i, subTask := range subTasks {
		nodeID := fmt.Sprintf("task_%d", subTask.ID)
		node := &domain.DAGNode{
			ID:        nodeID,
			SubTask:   subTask,
			DependsOn: subTask.Dependencies,
			Status:    subTask.Status,
			Metadata: map[string]interface{}{
				"index":           i,
				"execution_order": subTask.ExecutionOrder,
				"can_parallel":    subTask.CanParallel,
			},
		}
		dag.Nodes = append(dag.Nodes, node)
		nodeMap[nodeID] = node
	}

	// 创建边
	for _, node := range dag.Nodes {
		for _, depID := range node.DependsOn {
			// 检查依赖是否存在
			if _, exists := nodeMap[depID]; !exists {
				logger.Warn("依赖节点不存在: %s -> %s", node.ID, depID)
				continue
			}

			edge := &domain.DAGEdge{
				From:     depID,
				To:       node.ID,
				DataFlow: "result", // 默认数据流类型
			}
			dag.Edges = append(dag.Edges, edge)
		}
	}

	// 验证DAG是否有环
	if de.hasCycles(dag) {
		return nil, fmt.Errorf("依赖图中存在循环依赖")
	}

	logger.Info("DAG构建完成: %d个节点, %d条边", len(dag.Nodes), len(dag.Edges))
	return dag, nil
}

// topologicalSort 拓扑排序
func (de *DAGEngine) topologicalSort(dag *domain.DependencyGraph) ([]string, error) {
	// 计算入度
	inDegree := make(map[string]int)
	for _, node := range dag.Nodes {
		inDegree[node.ID] = 0
	}
	for _, edge := range dag.Edges {
		inDegree[edge.To]++
	}

	// 初始化队列：入度为0的节点
	queue := make([]string, 0)
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	// 拓扑排序
	topology := make([]string, 0, len(dag.Nodes))
	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		topology = append(topology, nodeID)

		// 减少依赖节点的入度
		for _, edge := range dag.Edges {
			if edge.From == nodeID {
				inDegree[edge.To]--
				if inDegree[edge.To] == 0 {
					queue = append(queue, edge.To)
				}
			}
		}
	}

	// 检查是否所有节点都被排序
	if len(topology) != len(dag.Nodes) {
		return nil, fmt.Errorf("图中存在环，无法进行拓扑排序")
	}

	dag.Topology = topology
	logger.Info("拓扑排序完成: %v", topology)
	return topology, nil
}

// executeTopology 按拓扑顺序执行任务
func (de *DAGEngine) executeTopology(ctx context.Context, dag *domain.DependencyGraph, topology []string, task *domain.Task) (string, error) {
	// 创建节点映射
	nodeMap := make(map[string]*domain.DAGNode)
	for _, node := range dag.Nodes {
		nodeMap[node.ID] = node
	}

	// 结果映射
	resultMap := make(map[string]string)
	resultMutex := &sync.RWMutex{}
	var executionErrors []error

	// 按拓扑顺序分批执行
	for _, nodeID := range topology {
		node := nodeMap[nodeID]

		// 等待依赖完成
		for _, depID := range node.DependsOn {
			for {
				resultMutex.RLock()
				_, hasResult := resultMap[depID]
				resultMutex.RUnlock()

				if hasResult {
					break
				}
				// 简单的等待机制，实际应该用通道或条件变量
			}
		}

		// 执行当前任务
		logger.Info("开始执行节点 %s: %s", node.ID, node.SubTask.Description)

		// 收集依赖结果
		inputData := make(map[string]interface{})
		if node.SubTask.Input != nil {
			for k, v := range node.SubTask.Input {
				inputData[k] = v
			}
		}

		// 添加依赖结果到输入
		for _, depID := range node.DependsOn {
			resultMutex.RLock()
			if result, ok := resultMap[depID]; ok {
				inputData[fmt.Sprintf("dep_%s", depID)] = result
			}
			resultMutex.RUnlock()
		}

		// 执行子任务
		err := de.executeSubTask(ctx, node.SubTask, inputData)
		if err != nil {
			executionErrors = append(executionErrors, fmt.Errorf("节点 %s 执行失败: %w", nodeID, err))
			node.Status = domain.StatusFailed
			node.SubTask.Status = domain.StatusFailed
			node.SubTask.Error = err.Error()

			// 任务失败，终止后续执行
			logger.Error("节点执行失败，终止DAG执行")
			return "", fmt.Errorf("节点 %s 执行失败: %w", nodeID, err)
		}

		// 保存结果
		resultMutex.Lock()
		resultMap[nodeID] = node.SubTask.Output
		resultMutex.Unlock()

		node.Status = domain.StatusCompleted
		node.SubTask.Status = domain.StatusCompleted

		logger.Info("节点 %s 执行成功", node.ID)
	}

	// 所有任务完成，聚合最终结果
	if len(executionErrors) > 0 {
		return "", fmt.Errorf("%d个节点执行失败", len(executionErrors))
	}

	// 获取最后一个节点的结果作为最终结果
	lastNodeID := topology[len(topology)-1]
	resultMutex.RLock()
	finalResult := resultMap[lastNodeID]
	resultMutex.RUnlock()

	return finalResult, nil
}

// executeSubTask 执行单个子任务
func (de *DAGEngine) executeSubTask(ctx context.Context, subTask *domain.SubTask, inputData map[string]interface{}) error {
	// 创建任务上下文
	taskCtx := de.contextManager.CreateTaskContext(
		subTask.ID,
		"You are a helpful assistant that executes tasks efficiently.",
	)
	subTask.Context = taskCtx

	// 构建任务提示
	prompt := fmt.Sprintf("请执行以下任务：\n\n任务描述：%s\n\n", subTask.Description)

	if len(inputData) > 0 {
		prompt += "输入信息：\n"
		for key, value := range inputData {
			if strings.HasPrefix(key, "dep_") {
				prompt += fmt.Sprintf("- 依赖任务的结果：\n%s\n", value)
			} else {
				prompt += fmt.Sprintf("- %s: %v\n", key, value)
			}
		}
	}

	prompt += "\n请详细完成该任务，并返回结果。"

	// 添加用户消息到上下文
	de.contextManager.AddMessage(subTask.ID, "user", prompt, len(prompt))

	// 执行任务
	messages := []domain.Message{
		{Role: "system", Content: taskCtx.SystemPrompt},
		{Role: "user", Content: prompt},
	}

	response := de.taskChat.Chat(ctx, messages)

	// 添加助手响应到上下文
	de.contextManager.AddMessage(subTask.ID, "assistant", response, len(response))

	// 更新子任务状态
	subTask.Status = domain.StatusCompleted
	subTask.Output = response

	return nil
}

// hasCycles 检测DAG是否有环
func (de *DAGEngine) hasCycles(dag *domain.DependencyGraph) bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for _, node := range dag.Nodes {
		if !visited[node.ID] {
			if de.detectCycleDFS(node.ID, dag, visited, recStack) {
				return true
			}
		}
	}
	return false
}

// detectCycleDFS 深度优先搜索检测环
func (de *DAGEngine) detectCycleDFS(nodeID string, dag *domain.DependencyGraph, visited, recStack map[string]bool) bool {
	// 查找节点
	var node *domain.DAGNode
	for _, n := range dag.Nodes {
		if n.ID == nodeID {
			node = n
			break
		}
	}
	if node == nil {
		return false
	}

	visited[nodeID] = true
	recStack[nodeID] = true

	// 检查所有依赖
	for _, depID := range node.DependsOn {
		if !visited[depID] {
			if de.detectCycleDFS(depID, dag, visited, recStack) {
				return true
			}
		} else if recStack[depID] {
			return true
		}
	}

	recStack[nodeID] = false
	return false
}