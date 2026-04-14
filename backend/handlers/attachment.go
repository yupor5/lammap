package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"quotepro-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func ListAttachments(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		productID, _ := strconv.Atoi(c.Query("productId"))
		quoteID, _ := strconv.Atoi(c.Query("quoteId"))

		query := db.Model(&models.Attachment{}).Where("user_id = ?", userID)
		if productID > 0 {
			var p models.Product
			if err := db.First(&p, productID).Error; err != nil || p.UserID != userID {
				Success(c, []models.Attachment{})
				return
			}
			query = query.Where("product_id = ?", productID)
		}
		if quoteID > 0 {
			var q models.Quote
			if err := db.First(&q, quoteID).Error; err != nil || q.UserID != userID {
				Success(c, []models.Attachment{})
				return
			}
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

		if productID > 0 {
			var p models.Product
			if err := db.First(&p, productID).Error; err != nil || p.UserID != userID {
				Error(c, http.StatusBadRequest, "产品不存在或无权访问")
				return
			}
		}
		if quoteID > 0 {
			var q models.Quote
			if err := db.First(&q, quoteID).Error; err != nil || q.UserID != userID {
				Error(c, http.StatusBadRequest, "报价不存在或无权访问")
				return
			}
		}

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

		ext := strings.ToLower(filepath.Ext(file.Filename))
		filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
		dst := filepath.Join(uploadDir, filename)

		orig, err := saveMultipartFileValidated(file, dst)
		if err != nil {
			_ = os.Remove(dst)
			Error(c, http.StatusBadRequest, err.Error())
			return
		}

		st, _ := os.Stat(dst)
		sz := file.Size
		if st != nil {
			sz = st.Size()
		}

		attachment := models.Attachment{
			UserID:    userID,
			ProductID: uint(productID),
			QuoteID:   uint(quoteID),
			FileName:  orig,
			FilePath:  "/" + dst,
			FileSize:  sz,
			FileType:  ext,
		}

		if err := db.Create(&attachment).Error; err != nil {
			Error(c, http.StatusInternalServerError, "保存附件记录失败")
			return
		}

		// 回写产品附件数，供列表展示
		if productID > 0 {
			var cnt int64
			db.Model(&models.Attachment{}).Where("user_id = ? AND product_id = ?", userID, productID).Count(&cnt)
			db.Model(&models.Product{}).Where("id = ? AND user_id = ?", productID, userID).Update("attachments", int(cnt))
		}

		Success(c, attachment)
	}
}

func DeleteAttachment(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")
		var attachment models.Attachment
		if err := db.First(&attachment, id).Error; err != nil {
			Error(c, http.StatusNotFound, "附件不存在")
			return
		}
		if attachment.UserID != userID {
			Error(c, http.StatusNotFound, "附件不存在")
			return
		}

		if len(attachment.FilePath) > 1 {
			_ = os.Remove(attachment.FilePath[1:])
		}

		if err := db.Delete(&attachment).Error; err != nil {
			Error(c, http.StatusInternalServerError, "删除失败")
			return
		}

		// 回写产品附件数（删除后）
		if attachment.ProductID > 0 {
			var cnt int64
			db.Model(&models.Attachment{}).Where("user_id = ? AND product_id = ?", userID, attachment.ProductID).Count(&cnt)
			db.Model(&models.Product{}).Where("id = ? AND user_id = ?", attachment.ProductID, userID).Update("attachments", int(cnt))
		}
		Success(c, nil)
	}
}
