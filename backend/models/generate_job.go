package models

import (
	"time"

	"gorm.io/gorm"
)

type GenerateJobStatus string

const (
	GenerateJobQueued    GenerateJobStatus = "queued"
	GenerateJobRunning   GenerateJobStatus = "running"
	GenerateJobSucceeded GenerateJobStatus = "succeeded"
	GenerateJobFailed    GenerateJobStatus = "failed"
)

// GenerateJob 用于异步 AI 生成报价的后台任务。
// ResultJSON 存 normalizeGenerateResponse 后的结果 JSON（前端可直接映射为 Quote）。
type GenerateJob struct {
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

