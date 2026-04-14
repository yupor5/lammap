package models

import (
	"time"

	"gorm.io/gorm"
)

type Template struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID    uint           `json:"userId"`
	Name      string         `gorm:"size:255" json:"name"`
	Category  string         `gorm:"size:50" json:"category"`
	Language  string         `gorm:"size:10;index" json:"language"`
	Content   string         `gorm:"type:text" json:"content"`
	// Source: user（手工） / ai（AI 生成） / system（预置，预留）
	Source string `gorm:"size:20;default:user;index" json:"source"`
}
