package task

import (
	"orange-agent/agent"
	"orange-agent/agent/interfaces"
)

// SimpleAgentManager 简单的Agent管理器实现
type SimpleAgentManager struct {
	defaultAgent interfaces.Agent
}

// NewSimpleAgentManager 创建新的简单Agent管理器
func NewSimpleAgentManager() *SimpleAgentManager {
	return &SimpleAgentManager{
		defaultAgent: agent.NewAgent(),
	}
}

// GetDefaultAgent 获取默认Agent
func (sam *SimpleAgentManager) GetDefaultAgent() (interfaces.Agent, error) {
	if sam.defaultAgent == nil {
		return nil, ErrAgentNotFound
	}
	return sam.defaultAgent, nil
}

// GetAgentByName 根据名称获取Agent（这里简化实现，只返回默认Agent）
func (sam *SimpleAgentManager) GetAgentByName(name string) (interfaces.Agent, error) {
	return sam.GetDefaultAgent()
}
