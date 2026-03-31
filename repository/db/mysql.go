package db

import (
	"orange-agent/config"
	"orange-agent/domain"
	"orange-agent/utils/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Mysql struct {
	DB     *gorm.DB
	config *config.DatabaseConfig
}

func NewMysql(config *config.DatabaseConfig) (*Mysql, error) {
	db := &Mysql{
		config: config,
	}
	log := logger.GetLogger()
	coon, err := gorm.Open(mysql.Open(buildDsn(config)), &gorm.Config{})
	if err != nil {
		log.Error("数据库连接失败:%s", err.Error())
		return nil, err
	}
	log.Info("数据库连接成功")
	db.DB = coon
	db.config = config
	Migrate(db) // 迁移
	return db, nil
}

func buildDsn(config *config.DatabaseConfig) string {
	return config.Username + ":" + config.Password + "@tcp(" + config.Host + ":" + config.Port + ")/" + config.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
}

// 迁移
func Migrate(mysql *Mysql) {
	mysql.DB.AutoMigrate(
		&domain.Task{},
		&domain.SubTask{},
		&domain.TaskResult{},
	)
}
