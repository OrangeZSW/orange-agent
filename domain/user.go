package domain

import (
	"gorm.io/gorm"
)

var (
	NORMAL = "normal"
	TASK   = "task"
)

type User struct {
	gorm.Model
	Name       string `json:"name"`
	TelegramId uint   `json:"telegram_id"`
	ModelName  string
	ChainMode  string `json:"chain_mode" gorm:"default:normal"`
}
