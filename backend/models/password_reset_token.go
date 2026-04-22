package models

import (
	"time"

	"gorm.io/gorm"
)

// PasswordResetToken 忘记密码流程：仅存 token，需配合邮件或线下传递；单 token 一次性使用。
type PasswordResetToken struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Email     string    `gorm:"size:255;index" json:"email"`
	Token     string    `gorm:"size:64;uniqueIndex" json:"-"`
	ExpiresAt time.Time `json:"expiresAt"`
	Used      bool      `gorm:"default:false" json:"used"`
}
