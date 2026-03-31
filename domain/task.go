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
