package mysql

import (
	"errors"
	"orange-agent/domain"

	"gorm.io/gorm"
)

type MemorySql struct {
	db *gorm.DB
}

func NewMemorySql() *MemorySql {
	return &MemorySql{
		db: GetDB(),
	}
}

func (m *MemorySql) CreateMemory(memory *domain.Memory) error {
	return m.db.Create(memory).Error
}

// get by user id
func (m *MemorySql) GetMemoryByUserId(userId uint) (*[]domain.Memory, error) {
	var memories []domain.Memory
	err := m.db.Where("user_id = ?", userId).Find(&memories).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &memories, err
}

// update
func (m *MemorySql) UpdateMemory(memory *domain.Memory) error {
	return m.db.Save(memory).Error
}

// getMemoryByIdAndSize
func (m *MemorySql) GetMemoryByIdAndSize(memoryId uint, size int) (*[]domain.Memory, error) {
	var memories []domain.Memory
	err := m.db.Where("id = ?", memoryId).Limit(size).Find(&memories).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &memories, err
}
