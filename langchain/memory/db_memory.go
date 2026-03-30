package memory

import (
	"orange-agent/domain"
	factory "orange-agent/repository/factory"
	"orange-agent/utils/logger"
)

type DBMemoryManager struct {
	repo factory.Factory
	log  *logger.Logger
}

func NewDBMemoryManager() *DBMemoryManager {
	return &DBMemoryManager{
		repo: *factory.NewFactory(),
		log:  logger.GetLogger(),
	}
}

func (m *DBMemoryManager) GetMemory(userID uint, limit int) ([]domain.Memory, error) {
	memories, err := m.repo.MemoryRepo.GetMemoryByUserIdAndLimit(userID, limit)
	if err != nil {
		m.log.Error("获取用户 %d 记忆失败: %v", userID, err)
		return nil, err
	}

	m.log.Debug("加载用户ID %d 记忆：%d 条", userID, len(memories))
	return memories, nil
}

func (m *DBMemoryManager) SaveMemory(memory *domain.Memory) error {
	if err := m.repo.MemoryRepo.CreateMemory(memory); err != nil {
		m.log.Error("保存记忆失败: %v", err)
		return err
	}

	m.log.Debug("成功保存记忆，用户ID: %d", memory.UserId)
	return nil
}
