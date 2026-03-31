package factory

import (
	"orange-agent/repository"
	orange_grom "orange-agent/repository/grom"
	"orange-agent/repository/resource"
	"sync"
)

var (
	repoFactory *Factory
	once        sync.Once
)

type Factory struct {
	dbResource resource.DataResource

	AgentCallRecordRepo repository.AgentCallRecordRepository
	AgentConfigRepo     repository.AgentConfigRepository
	UserRepo            repository.UserRepository
	MemoryRepo          repository.MemoryRepository
	TaskRepo            repository.TaskRepository
	SubTaskRepo         repository.SubTaskRepository
	TaskResultRepo      repository.TaskResultRepository
}

func NewFactory() *Factory {
	once.Do(func() {
		Factory := &Factory{
			dbResource: *resource.GetDataResource(),
		}
		gormDB := resource.GetDataResource().Mysql.DB

		if gormDB != nil {
			Factory.AgentCallRecordRepo = orange_grom.NewAgentCallRecordRepository(gormDB)
			Factory.AgentConfigRepo = orange_grom.NewAgentConfigRepository(gormDB)
			Factory.UserRepo = orange_grom.NewUserRepository(gormDB)
			Factory.MemoryRepo = orange_grom.NewMemoryRepository(gormDB)
		}
		repoFactory = Factory
	})

	return repoFactory
}
