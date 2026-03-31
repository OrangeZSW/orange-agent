package resource

import (
	"orange-agent/repository"
	"orange-agent/utils/logger"
	"sync"
)

var (
	dataResource *DataResource
	once         sync.Once
)

type Resource interface {
	InitRepo() (*repository.Repositories, error)
}

type DataResource struct {
	Resource     map[string]Resource
	log          *logger.Logger
	repositories *repository.Repositories
}

func GetDataResource() *DataResource {
	once.Do(func() {
		dataResource = &DataResource{
			Resource: make(map[string]Resource),
			log:      logger.GetLogger(),
		}
	})
	return dataResource
}

func (r *DataResource) Add(resource Resource, name string) {
	r.Resource[name] = resource
	repositories, err := resource.InitRepo()
	if err != nil {
		r.log.Error("InitRepo error: %v", err)
		return
	}
	r.repositories = repositories
}

func GetRepositories() *repository.Repositories {
	return GetDataResource().repositories
}
