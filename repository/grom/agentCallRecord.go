package gorm

import (
	"errors"
	"orange-agent/domain"
	"orange-agent/repository"

	"gorm.io/gorm"
)

type agentCallRecordRepository struct {
	db *gorm.DB
}

// 构造函数：接收 db 参数，便于依赖注入
func NewAgentCallRecordRepository(db *gorm.DB) repository.AgentCallRecordRepository {
	return &agentCallRecordRepository{
		db: db,
	}
}

// CreateAgentCallRecord 创建调用记录
func (r *agentCallRecordRepository) CreateAgentCallRecord(record *domain.CallRecord) error {
	if record == nil {
		return errors.New("record cannot be nil")
	}
	return r.db.Create(record).Error
}

// GetAgentCallRecordByAgentName 根据 Agent 名称获取所有调用记录
func (r *agentCallRecordRepository) GetAgentCallRecordByAgentName(agentName string) ([]domain.CallRecord, error) {
	var records []domain.CallRecord

	// 使用指针传递切片，Find 会自动填充
	err := r.db.Where("agent_name = ?", agentName).
		Order("created_at DESC"). // 按时间倒序
		Find(&records).Error

	// 如果没有找到记录，返回空切片而不是 nil
	if err != nil {
		return []domain.CallRecord{}, err
	}
	return records, nil
}
