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
	rows, err := r.db.Exec(query, args...).Rows()
	if err != nil {
		return nil, err
	}
	return rows, err
}

func (r *sqlQueryRepository) Execute(query string, args ...interface{}) (tx *gorm.DB) {
	return r.db.Raw(query, args...)

}
