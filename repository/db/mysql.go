package db

import (
	"orange-agent/domain"
	"orange-agent/repository"
	"orange-agent/utils/logger"

	mysql_gorm "orange-agent/repository/gorm"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Mysql struct {
	db     *gorm.DB
	config *domain.DatabaseConfig
}

func InitMysql(config *domain.DatabaseConfig) *Mysql {
	db := &Mysql{
		config: config,
	}
	log := logger.GetLogger()
	coon, err := gorm.Open(mysql.Open(buildDsn(config)), &gorm.Config{})
	if err != nil {
		log.Error("数据库连接失败:%s", err.Error())
		return nil
	}
	log.Info("数据库连接成功")
	db.db = coon
	db.config = config
	Migrate(db) // 迁移
	return db
}

func buildDsn(config *domain.DatabaseConfig) string {
	return config.Username + ":" + config.Password + "@tcp(" + config.Host + ":" + config.Port + ")/" + config.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
}

// 迁移
func Migrate(mysql *Mysql) {
	mysql.db.AutoMigrate(
		&domain.Task{},
		&domain.SubTask{},
		&domain.TaskResult{},
	)
}

func (m *Mysql) InitRepo() (*repository.Repositories, error) {
	return &repository.Repositories{
		Task:            mysql_gorm.NewTaskRepository(m.db),
		SubTask:         mysql_gorm.NewSubTaskRepository(m.db),
		TaskResult:      mysql_gorm.NewTaskResultRepository(m.db),
		User:            mysql_gorm.NewUserRepository(m.db),
		AgentConfig:     mysql_gorm.NewAgentConfigRepository(m.db),
		Memory:          mysql_gorm.NewMemoryRepository(m.db),
		AgentCallRecord: mysql_gorm.NewAgentCallRecordRepository(m.db),
	}, nil
}
