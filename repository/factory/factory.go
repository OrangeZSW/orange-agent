package factory

import (
	"orange-agent/repository"
	dataresource "orange-agent/repository/data_resource"
	orange_grom "orange-agent/repository/grom"
	"sync"
)

var (
	repoFactory *Factory
	once        sync.Once
)

type Factory struct {
	dbResource dataresource.DataResource

	AgentCallRecordRepo repository.AgentCallRecordRepository
	AgentConfigRepo     repository.AgentConfigRepository
	UserRepo            repository.UserRepository
	MemoryRepo          repository.MemoryRepository
}

func NewFactory() *Factory {
	once.Do(func() {
		Factory := &Factory{
			dbResource: *dataresource.GetDataResource(),
		}
		gormDB := dataresource.GetDataResource().Mysql.DB

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
