package gorm

import (
	"orange-agent/domain"
	"orange-agent/repository"

	"gorm.io/gorm"
)

type taskResultRepository struct {
	db *gorm.DB
}

// CreateTaskResult(taskResult *domain.TaskResult) error
// 	GetTaskResultBySubTaskId(subTaskId uint) (*domain.TaskResult, error)
// 	UpdateTaskResult(taskResult *domain.TaskResult) error
// 	GetTaskResultById(id uint) (*domain.TaskResult, error)

func NewTaskResultRepository(db *gorm.DB) repository.TaskResultRepository {
	return &taskResultRepository{
		db: db,
	}
}

func (r *taskResultRepository) CreateTaskResult(taskResult *domain.TaskResult) error {
	return r.db.Create(taskResult).Error
}

func (r *taskResultRepository) GetTaskResultBySubTaskId(subTaskId uint) (*domain.TaskResult, error) {
	var taskResult domain.TaskResult
	err := r.db.Where("sub_task_id = ?", subTaskId).First(&taskResult).Error
	if err != nil {
		return nil, err
	}
	return &taskResult, nil
}

func (r *taskResultRepository) UpdateTaskResult(taskResult *domain.TaskResult) error {
	return r.db.Save(taskResult).Error
}

func (r *taskResultRepository) GetTaskResultById(id uint) (*domain.TaskResult, error) {
	var taskResult domain.TaskResult
	err := r.db.First(&taskResult, id).Error
	if err != nil {
		return nil, err
	}
	return &taskResult, nil
}
