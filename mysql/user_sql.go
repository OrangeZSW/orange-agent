package mysql

import (
	"errors"
	"orange-agent/domain"

	"gorm.io/gorm"
)

type UserSql struct {
	db *gorm.DB
}

func NewUserSql() *UserSql {
	return &UserSql{
		db: GetDB(),
	}
}

func (u *UserSql) CreateUser(user *domain.User) error {
	return u.db.Create(user).Error
}

// get by telegram id
func (u *UserSql) GetUserByTelegramId(telegramId int64) (*domain.User, error) {
	var user domain.User
	err := u.db.Where("telegram_id = ?", telegramId).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &user, err
}

// update user model name
func (u *UserSql) UpdateUserModelName(telegramId int64, modelName string) error {
	return u.db.Model(&domain.User{}).Where("telegram_id = ?", telegramId).Update("model_name", modelName).Error
}
