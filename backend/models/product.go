package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	UserID       uint           `json:"userId"`
	Name         string         `gorm:"size:255" json:"name"`
	SKU          string         `gorm:"size:100" json:"sku"`
	Category     string         `gorm:"size:100" json:"category"`
	Description  string         `gorm:"type:text" json:"description"`
	Material     string         `gorm:"size:255" json:"material"`
	Size         string         `gorm:"size:255" json:"size"`
	Color        string         `gorm:"size:100" json:"color"`
	Process      string         `gorm:"size:255" json:"process"`
	Packaging    string         `gorm:"size:255" json:"packaging"`
	Price        float64        `json:"price"`
	MOQ          int            `json:"moq"`
	LeadTime     string         `gorm:"size:100" json:"leadTime"`
	PaymentTerms string         `gorm:"size:255" json:"paymentTerms"`
	Attachments  int            `json:"attachments"`
}
