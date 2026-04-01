package gorm

import (
	"database/sql"
	"orange-agent/repository"

	"gorm.io/gorm"
)

type sqlQueryRepository struct {
	db *gorm.DB
}

func NewSqlQueryRepository(db *gorm.DB) repository.SqlQuery {
	return &sqlQueryRepository{db: db}
}

func (r *sqlQueryRepository) ExecuteRows(query string, args ...interface{}) (*sql.Rows, error) {
	// 修复：查询操作应该使用Raw而不是Exec
	rows, err := r.db.Raw(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	return rows, err
}

func (r *sqlQueryRepository) Execute(query string, args ...interface{}) (tx *gorm.DB) {
	// 修复：写操作应该使用Exec而不是Raw
	return r.db.Exec(query, args...)

}