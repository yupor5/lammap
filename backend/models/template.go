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
	Content   string         `gorm:"type:text" json:"content"`
}
