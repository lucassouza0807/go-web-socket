package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	Avatar   string    `gorm:"size:255" json:"avatar"`
	UserId   string    `gorm:"size:255;unique" json:"user_id"`
	Username string    `gorm:"size:255;unique" json:"username"`
	Name     string    `gorm:"size:150" json:"name"`
	Password string    `gorm:"size:150"`
	Messages []Message `gorm:"foreignKey:UserID"`
}

type Message struct {
	ID        uint           `gorm:"primaryKey"`
	UserID    uint           `gorm:"not null"`
	User      User           `gorm:"constraint:OnDelete:CASCADE;"`
	Content   string         `gorm:"type:text;not null"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
