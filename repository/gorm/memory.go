package gorm

import (
	"errors"
	"orange-agent/domain"
	"orange-agent/repository"

	"gorm.io/gorm"
)

type memoryRepository struct {
	db *gorm.DB
}

// 构造函数：接收 db 参数，便于依赖注入
func NewMemoryRepository(db *gorm.DB) repository.MemoryRepository {
	return &memoryRepository{
		db: db,
	}
}

// CreateMemory 创建记忆
func (r *memoryRepository) CreateMemory(memory *domain.Memory) error {
	if memory == nil {
		return errors.New("memory cannot be nil")
	}
	return r.db.Create(memory).Error
}

// GetMemoryByUserId 根据用户 ID 获取记忆
func (r *memoryRepository) GetMemoryByUserId(userId uint) ([]domain.Memory, error) {
	if userId == 0 {
		return []domain.Memory{}, errors.New("userId cannot be zero")
	}

	var memories []domain.Memory
	err := r.db.Where("user_id = ?", userId).Find(&memories).Error
	if err != nil {
		return []domain.Memory{}, err
	}
	return memories, nil
}

// UpdateMemory 更新记忆
func (r *memoryRepository) UpdateMemory(memory *domain.Memory) error {
	if memory == nil {
		return errors.New("memory cannot be nil")
	}
	return r.db.Save(memory).Error
}

// GetMemoryByUserIdAndLimit 根据用户 ID 和限制获取记忆
func (r *memoryRepository) GetMemoryByUserIdAndLimit(userId uint, limit int) ([]domain.Memory, error) {
	if userId == 0 {
		return []domain.Memory{}, errors.New("userId cannot be zero")
	}
	if limit <= 0 {
		return []domain.Memory{}, errors.New("limit must be positive")
	}

	var memories []domain.Memory
	err := r.db.Where("user_id = ?", userId).Order("created_at DESC").Limit(limit).Find(&memories).Error
	if err != nil {
		return []domain.Memory{}, err
	}
	return memories, nil
}
