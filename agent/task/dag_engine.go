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
	dag, err := de.BuildDAG(task.Subtasks)
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

// BuildDAG 构建有向无环图
func (de *DAGEngine) BuildDAG(subTasks []*domain.SubTask) (*domain.DependencyGraph, error) {
	dag := &domain.DependencyGraph{
		Nodes:    make([]*domain.DAGNode, 0, len(subTasks)),
		Edges:    make([]*domain.DAGEdge, 0),
		Metadata: make(map[string]any),
	}

	// 创建节点映射
	nodeMap := make(map[string]*domain.DAGNode)
	for i, subTask := range subTasks {
		// 修复点1：检查子任务ID是否为默认值0，避免重复使用默认值
		if subTask.ID == 0 {
			return nil, fmt.Errorf("子任务ID不能为默认值0，请确保每个子任务ID全局唯一，子任务描述: %s", subTask.Description)
		}

		nodeID := fmt.Sprintf("task_%d", subTask.ID)

		// 修复点2：新增重复nodeID校验
		if _, exists := nodeMap[nodeID]; exists {
			return nil, fmt.Errorf("检测到重复的节点ID: %s，子任务ID: %d，请确保子任务ID全局唯一", nodeID, subTask.ID)
		}

		// 修复点3：新增自依赖校验
		for _, depID := range subTask.Dependencies {
			if depID == nodeID {
				return nil, fmt.Errorf("节点 %s 存在自依赖，任务不能依赖自身，子任务描述: %s", nodeID, subTask.Description)
			}
		}

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
		logger.Debug("创建DAG节点: %s, 依赖列表: %v", nodeID, subTask.Dependencies)
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
			logger.Debug("创建DAG边: %s -> %s", depID, node.ID)
		}
	}

	// 验证DAG是否有环
	if de.hasCycles(dag) {
		// 打印所有节点依赖关系方便排查
		logger.Error("DAG构建失败，存在循环依赖，所有节点依赖关系如下:")
		for _, node := range dag.Nodes {
			logger.Error("节点 %s 依赖: %v", node.ID, node.DependsOn)
		}
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

	logger.Debug("节点入度统计: %v", inDegree)

	// 初始化队列：入度为0的节点
	queue := make([]string, 0)
	for nodeID, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, nodeID)
		}
	}

	logger.Debug("初始入度为0的节点: %v", queue)

	// 拓扑排序
	topology := make([]string, 0, len(dag.Nodes))
	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]
		topology = append(topology, nodeID)

		logger.Debug("处理拓扑节点: %s", nodeID)

		// 减少依赖节点的入度
		for _, edge := range dag.Edges {
			if edge.From == nodeID {
				inDegree[edge.To]--
				logger.Debug("节点 %s 依赖的节点 %s 已处理，入度减为: %d", edge.To, nodeID, inDegree[edge.To])
				if inDegree[edge.To] == 0 {
					queue = append(queue, edge.To)
					logger.Debug("节点 %s 入度变为0，加入处理队列", edge.To)
				}
			}
		}
	}

	// 检查是否所有节点都被排序
	if len(topology) != len(dag.Nodes) {
		// 找出未被处理的节点
		var unprocessedNodes []string
		for nodeID := range inDegree {
			found := false
			for _, processedID := range topology {
				if nodeID == processedID {
					found = true
					break
				}
			}
			if !found {
				unprocessedNodes = append(unprocessedNodes, nodeID)
			}
		}
		logger.Error("拓扑排序失败，存在环，未处理的节点: %v, 剩余入度: %v", unprocessedNodes, inDegree)
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
		logger.Warn("检测环时找不到节点: %s", nodeID)
		return false
	}

	logger.Debug("开始检测节点环: %s, 依赖: %v", nodeID, node.DependsOn)

	visited[nodeID] = true
	recStack[nodeID] = true

	// 检查所有依赖
	for _, depID := range node.DependsOn {
		logger.Debug("节点 %s 检查依赖 %s", nodeID, depID)
		if !visited[depID] {
			if de.detectCycleDFS(depID, dag, visited, recStack) {
				// 打印环路径
				var cyclePath []string
				for n, inStack := range recStack {
					if inStack {
						cyclePath = append(cyclePath, n)
					}
				}
				logger.Error("检测到循环依赖! 环路径: %v -> %s", cyclePath, depID)
				return true
			}
		} else if recStack[depID] {
			// 找到环，打印具体的环信息
			var cyclePath []string
			for n, inStack := range recStack {
				if inStack {
					cyclePath = append(cyclePath, n)
				}
			}
			logger.Error("检测到循环依赖! 当前节点: %s, 依赖节点: %s 已在递归栈中, 完整环路径: %v -> %s", nodeID, depID, cyclePath, depID)
			return true
		}
	}

	recStack[nodeID] = false
	logger.Debug("节点 %s 检测完成，无环", nodeID)
	return false
}