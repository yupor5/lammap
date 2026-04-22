package models

import (
	"time"

	"gorm.io/gorm"
)

// ProductExampleJob 异步生成「产品示例」任务（产品资料库 AI 生成示例）。
type ProductExampleJob struct {
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

