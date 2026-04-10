package models

import (
	"time"

	"gorm.io/gorm"
)

// AttachmentGenerateJob 用于异步“AI 生成单个附件文件”，返回可访问 url。
type AttachmentGenerateJob struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID uint `gorm:"index" json:"userId"`

	Status   string `gorm:"size:20;index" json:"status"`
	ErrorMsg string `gorm:"type:text" json:"errorMsg"`

	RequestJSON string `gorm:"type:text" json:"requestJson"`
	ResultJSON  string `gorm:"type:text" json:"resultJson"`
}

