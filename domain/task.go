package domain

import (
	"gorm.io/gorm"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
)

type Task struct {
	gorm.Model
	SessionID   string     `json:"session_id"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	Subtasks    []*SubTask `json:"subtasks"`
	Result      string     `json:"result"`
}

type SubTask struct {
	gorm.Model
	Description string                 `json:"description"`
	Status      TaskStatus             `json:"status"`
	Context     *TaskContext           `json:"context" gorm:"serializer:json;type:json"`
	Input       map[string]interface{} `json:"input" gorm:"serializer:json;type:json"`
	Output      string                 `json:"output"`
	Error       string                 `json:"error"`
	TaskID      uint                   `json:"task_id"`
	Task        Task                   `gorm:"foreignKey:TaskID"`
	
	// 新增：显式依赖字段
	Dependencies []string `json:"dependencies" gorm:"serializer:json;type:json"`
	ExecutionOrder int    `json:"execution_order"` // 执行顺序，从0开始
	CanParallel   bool    `json:"can_parallel"`    // 是否可并行执行
	IsDAGNode     bool    `json:"is_dag_node"`     // 是否为DAG图节点
}

type TaskContext struct {
	SystemPrompt string                 `json:"system_prompt"`
	Messages     []Message              `json:"messages" gorm:"serializer:json;type:json"`
	TokenCount   int                    `json:"token_count"`
	Metadata     map[string]interface{} `json:"metadata" gorm:"serializer:json;type:json"`
}

type Message struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

type TaskResult struct {
	gorm.Model
	Success     bool    `json:"success"`
	Output      string  `json:"output"`
	TokenUsed   int     `json:"token_used"`
	ExecutionMs int64   `json:"execution_ms"`
	Error       string  `json:"error"`
	SubTaskID   uint    `json:"sub_task_id"`
	SubTask     SubTask `gorm:"foreignKey:SubTaskID"`
}

// DependencyGraph 依赖图结构
type DependencyGraph struct {
	Nodes    []*DAGNode          `json:"nodes"`
	Edges    []*DAGEdge          `json:"edges"`
	Topology []string            `json:"topology"` // 拓扑排序结果
	Metadata map[string]any      `json:"metadata"`
}

// DAGNode 有向无环图节点
type DAGNode struct {
	ID        string                 `json:"id"`        // 节点ID，通常是SubTask的ID或索引
	SubTask   *SubTask               `json:"sub_task"`  // 对应的子任务
	DependsOn []string               `json:"depends_on"` // 依赖的节点ID列表
	Status    TaskStatus             `json:"status"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// DAGEdge 有向无环图边
type DAGEdge struct {
	From     string `json:"from"`     // 源节点ID
	To       string `json:"to"`       // 目标节点ID
	DataFlow string `json:"data_flow"` // 数据流类型：result, output, input等
}