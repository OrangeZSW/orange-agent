package domain

import "gorm.io/gorm"

type Memory struct {
	gorm.Model
	UserId       uint
	User         User   `gorm:"foreignKey:UserId"`
	UserQuestion string `json:"user_question"`
	AgentAnswer  string `json:"agent_answer"`
}
