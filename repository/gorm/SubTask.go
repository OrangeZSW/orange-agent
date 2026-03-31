package gorm

import (
	"orange-agent/domain"
	"orange-agent/repository"

	"gorm.io/gorm"
)

type subTaskRepository struct {
	db *gorm.DB
}

func NewSubTaskRepository(db *gorm.DB) repository.SubTaskRepository {
	return &subTaskRepository{
		db: db,
	}
}

func (r *subTaskRepository) CreateSubTask(subTask *domain.SubTask) error {
	return r.db.Create(subTask).Error
}

// 修改方法名以匹配接口
func (r *subTaskRepository) GetSubTaskByTaskId(taskId uint) ([]domain.SubTask, error) {
	var subTasks []domain.SubTask // 改为值类型切片
	err := r.db.Where("task_id = ?", taskId).Find(&subTasks).Error
	if err != nil {
		return nil, err
	}
	return subTasks, nil
}

func (r *subTaskRepository) UpdateSubTask(subTask *domain.SubTask) error {
	return r.db.Save(subTask).Error
}

func (r *subTaskRepository) GetSubTaskById(id uint) (*domain.SubTask, error) {
	var subTask domain.SubTask
	err := r.db.First(&subTask, id).Error
	if err != nil {
		return nil, err
	}
	return &subTask, nil
}
