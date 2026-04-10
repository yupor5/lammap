package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"quotepro-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func ListAttachments(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID, _ := strconv.Atoi(c.Query("productId"))
		quoteID, _ := strconv.Atoi(c.Query("quoteId"))

		query := db.Model(&models.Attachment{})
		if productID > 0 {
			query = query.Where("product_id = ?", productID)
		}
		if quoteID > 0 {
			query = query.Where("quote_id = ?", quoteID)
		}

		var attachments []models.Attachment
		query.Order("created_at DESC").Find(&attachments)
		Success(c, attachments)
	}
}

func UploadAttachment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		productID, _ := strconv.Atoi(c.PostForm("productId"))
		quoteID, _ := strconv.Atoi(c.PostForm("quoteId"))

		file, err := c.FormFile("file")
		if err != nil {
			Error(c, http.StatusBadRequest, "请选择文件")
			return
		}

		uploadDir := filepath.Join("uploads", "attachments", time.Now().Format("2006-01"))
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			Error(c, http.StatusInternalServerError, "创建目录失败")
			return
		}

		ext := filepath.Ext(file.Filename)
		filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
		dst := filepath.Join(uploadDir, filename)

		if err := c.SaveUploadedFile(file, dst); err != nil {
			Error(c, http.StatusInternalServerError, "保存文件失败")
			return
		}

		attachment := models.Attachment{
			UserID:    userID,
			ProductID: uint(productID),
			QuoteID:   uint(quoteID),
			FileName:  file.Filename,
			FilePath:  "/" + dst,
			FileSize:  file.Size,
			FileType:  ext,
		}

		if err := db.Create(&attachment).Error; err != nil {
			Error(c, http.StatusInternalServerError, "保存附件记录失败")
			return
		}

		Success(c, attachment)
	}
}

func DeleteAttachment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var attachment models.Attachment
		if err := db.First(&attachment, id).Error; err != nil {
			Error(c, http.StatusNotFound, "附件不存在")
			return
		}

		os.Remove(attachment.FilePath[1:])

		if err := db.Delete(&attachment).Error; err != nil {
			Error(c, http.StatusInternalServerError, "删除失败")
			return
		}
		Success(c, nil)
	}
}
