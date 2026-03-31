package gorm

import (
	"orange-agent/domain"
	"orange-agent/repository"

	"gorm.io/gorm"
)

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) repository.TaskRepository {
	return &taskRepository{
		db: db,
	}
}

func (r *taskRepository) CreateTask(task *domain.Task) error {
	return r.db.Create(task).Error
}

func (r *taskRepository) GetTaskById(id uint) (*domain.Task, error) {
	var task domain.Task
	err := r.db.Preload("SubTasks").First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) UpdateTask(task *domain.Task) error {
	return r.db.Save(task).Error
}
