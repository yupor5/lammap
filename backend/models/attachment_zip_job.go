package models

import (
	"time"

	"gorm.io/gorm"
)

// AttachmentZipJob 用于异步打包“邮件附件包”（zip），仅打包已存在 url 的附件。
type AttachmentZipJob struct {
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

