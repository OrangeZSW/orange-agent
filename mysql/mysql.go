package mysql

import (
	"orange-agent/config"
	"orange-agent/domain"
	"orange-agent/utils/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

type Mysql struct {
	db     *gorm.DB
	config *config.DatabaseConfig
}

func NewMysql(config *config.DatabaseConfig) {
	log := logger.GetLogger()
	coon, err := gorm.Open(mysql.Open(buildDsn(config)), &gorm.Config{})
	if err != nil {
		log.Error("数据库连接失败:%s", err.Error())
		return
	}
	log.Info("数据库连接成功")
	db = coon
	Migrate() // 迁移
}

func buildDsn(config *config.DatabaseConfig) string {
	return config.Username + ":" + config.Password + "@tcp(" + config.Host + ":" + config.Port + ")/" + config.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
}

func GetDB() *gorm.DB {
	return db
}

// 迁移
func Migrate() {
	db := GetDB()
	db.AutoMigrate(&domain.AgentConfig{}, &domain.CallRecord{},
		&domain.Memory{}, &domain.User{},
	)
}
