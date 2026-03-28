package domain

import (
	"gorm.io/gorm"
)

type AgentConfig struct {
	gorm.Model
	Name    string   `json:"name"`
	Token   string   `json:"token"`
	BaseUrl string   `json:"base_url"`
	Models  []string `gorm:"serializer:json" json:"models"`
}

// 调用记录
type CallRecord struct {
	gorm.Model
	UserID           uint         `json:"user_id"`
	AgentId          uint         `json:"agent_id"`
	AgentName        string       `json:"agent_name"`
	CompletionTokens int          `json:"completion_tokens"`
	PromptTokens     int          `json:"prompt_tokens"`
	TotalTokens      int          `json:"total_tokens"`
	AgentConfig      *AgentConfig `gorm:"foreignKey:AgentId"`
	User             *User        `gorm:"foreignKey:UserID"`
	MenmoryId        uint         `json:"menmory_id"`
	Memory           *Memory      `gorm:"foreignKey:MenmoryId"`
}
