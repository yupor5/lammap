package models

import (
	"time"

	"gorm.io/gorm"
)

// InquiryExampleJob 异步批量生成「询盘示例」任务（示例库 AI 生成）。
type InquiryExampleJob struct {
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
