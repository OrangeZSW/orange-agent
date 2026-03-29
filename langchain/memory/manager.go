package memory

import "orange-agent/domain"

type Manager interface {
	GetMemory(userID uint, limit int) ([]domain.Memory, error)
	SaveMemory(memory *domain.Memory) error
}
