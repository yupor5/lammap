package models

import (
	"time"

	"gorm.io/gorm"
)

type Quote struct {
	ID               uint           `gorm:"primarykey" json:"id"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
	UserID           uint           `json:"userId"`
	QuoteNumber      string         `gorm:"size:50;uniqueIndex" json:"quoteNumber"`
	CustomerName     string         `gorm:"size:255" json:"customerName"`
	Country          string         `gorm:"size:100" json:"country"`
	Currency         string         `gorm:"size:10" json:"currency"`
	DeliveryAddress  string         `gorm:"size:500" json:"deliveryAddress"`
	Status           string         `gorm:"size:50;default:草稿" json:"status"`
	TotalAmount      float64        `json:"totalAmount"`
	LeadTime         string         `gorm:"size:100" json:"leadTime"`
	Remarks          string         `gorm:"type:text" json:"remarks"`
	Terms            string         `gorm:"type:text" json:"terms"`
	RawRequirement   string         `gorm:"type:text" json:"rawRequirement"`
	ParsedParams     string         `gorm:"type:text" json:"parsedParams"`
	ReplyVersions    string         `gorm:"type:text" json:"replyVersions"`
	ConfirmationList string         `gorm:"type:text" json:"confirmationList"`
	AttachmentList   string         `gorm:"type:text" json:"attachmentList"`
	TemplateMeta     string         `gorm:"type:text" json:"templateMeta"`
	RenderedContents string         `gorm:"type:text" json:"renderedContents"`
	Items            []QuoteItem    `gorm:"foreignKey:QuoteID" json:"items"`
}

type QuoteItem struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	QuoteID     uint           `json:"quoteId"`
	ProductName string         `gorm:"size:255" json:"productName"`
	Model       string         `gorm:"size:100" json:"model"`
	Specs       string         `gorm:"size:500" json:"specs"`
	Quantity    int            `json:"quantity"`
	UnitPrice   float64        `json:"unitPrice"`
	TotalPrice  float64        `json:"totalPrice"`
	Remark      string         `gorm:"size:500" json:"remark"`
}
