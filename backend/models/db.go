package models

import (
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func InitDB(dbPath string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	return db
}

func AutoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&User{},
		&Product{},
		&Quote{},
		&QuoteItem{},
		&GenerateJob{},
		&AttachmentPackJob{},
		&AttachmentGenerateJob{},
		&AttachmentZipJob{},
		&InquiryExampleJob{},
		&ProductExampleJob{},
		&Template{},
		&Attachment{},
		&PasswordResetToken{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}
}
