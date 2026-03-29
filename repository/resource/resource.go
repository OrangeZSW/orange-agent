package resource

import (
	"orange-agent/config"
	"orange-agent/repository/db"
	"sync"
)

var (
	dataResource *DataResource
	once         sync.Once
)

type DataResource struct {
	Mysql *db.Mysql // 改为指针类型
}

func NewDataResource() *DataResource {
	once.Do(func() {
		dataResource = &DataResource{}
	})
	return dataResource
}

func (d *DataResource) InitMysql(config *config.DatabaseConfig) error {
	mysql, err := db.NewMysql(config)
	if err != nil {
		return err
	}
	dataResource.Mysql = mysql
	return nil
}

func GetDataResource() *DataResource {
	return NewDataResource()
}
