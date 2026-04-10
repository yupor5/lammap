package models

import (
	"time"

	"gorm.io/gorm"
)

type Attachment struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID    uint           `json:"userId"`
	ProductID uint           `json:"productId" gorm:"index"`
	QuoteID   uint           `json:"quoteId" gorm:"index"`
	FileName  string         `gorm:"size:255" json:"fileName"`
	FilePath  string         `gorm:"size:500" json:"filePath"`
	FileSize  int64          `json:"fileSize"`
	FileType  string         `gorm:"size:50" json:"fileType"`
}
