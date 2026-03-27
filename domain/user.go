package domain

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name       string `json:"name"`
	TelegramId uint   `json:"telegram_id"`
	ModelName  string
}
