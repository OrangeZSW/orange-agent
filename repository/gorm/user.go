package gorm

import (
	"errors"
	"orange-agent/domain"
	"orange-agent/repository"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// 构造函数：接收 db 参数，便于依赖注入
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{
		db: db,
	}
}

// GetUserById 根据 ID 获取用户
func (r *userRepository) GetUserById(id uint) (*domain.User, error) {
	if id == 0 {
		return nil, errors.New("id cannot be zero")
	}

	var user domain.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// CreateUser 创建用户
func (r *userRepository) CreateUser(user *domain.User) error {
	if user == nil {
		return errors.New("user cannot be nil")
	}
	return r.db.Create(user).Error
}

// GetUserByTelegramId 根据 Telegram ID 获取用户
func (r *userRepository) GetUserByTelegramId(telegramId int64) (*domain.User, error) {
	if telegramId == 0 {
		return nil, errors.New("telegramId cannot be zero")
	}

	var user domain.User
	err := r.db.Where("telegram_id = ?", telegramId).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUserModelName 更新用户模型名称
func (r *userRepository) UpdateUserModelName(telegramId int64, modelName string) error {
	if telegramId == 0 {
		return errors.New("telegramId cannot be zero")
	}
	if modelName == "" {
		return errors.New("modelName cannot be empty")
	}
	return r.db.Model(&domain.User{}).Where("telegram_id = ?", telegramId).Update("model_name", modelName).Error
}
