package mysql

import (
	"orange-agent/domain"

	"gorm.io/gorm"
)

type AgentCallRecordSql struct {
	db *gorm.DB
}

func NewAgentCallRecordSql() *AgentCallRecordSql {
	return &AgentCallRecordSql{
		db: GetDB(),
	}
}

// create
func (a *AgentCallRecordSql) CreateAgentCallRecord(agentCallRecord *domain.CallRecord) error {
	return a.db.Create(agentCallRecord).Error
}

// getByAgentName
func (a *AgentCallRecordSql) GetAgentCallRecordByAgentName(agentName string) (*[]domain.CallRecord, error) {
	var agentCallRecord domain.CallRecord
	return &[]domain.CallRecord{}, a.db.Where("agent_name = ?", agentName).Find(&agentCallRecord).Error
}
