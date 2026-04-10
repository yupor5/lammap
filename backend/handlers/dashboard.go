package handlers

import (
	"time"

	"quotepro-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DashboardStats(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		today := time.Now().Format("2006-01-02")
		weekAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")

		var todayCount, pendingCount, toSendCount, weekSentCount int64

		db.Model(&models.Quote{}).Where("user_id = ? AND DATE(created_at) = ?", userID, today).Count(&todayCount)
		db.Model(&models.Quote{}).Where("user_id = ? AND status = ?", userID, "待确认").Count(&pendingCount)
		db.Model(&models.Quote{}).Where("user_id = ? AND status = ?", userID, "草稿").Count(&toSendCount)
		db.Model(&models.Quote{}).Where("user_id = ? AND status = ? AND DATE(updated_at) >= ?", userID, "已发送", weekAgo).Count(&weekSentCount)

		Success(c, gin.H{
			"todayQuotes":   todayCount,
			"pendingParams": pendingCount,
			"toSend":        toSendCount,
			"weekSent":      weekSentCount,
		})
	}
}

func RecentQuotes(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		var quotes []models.Quote
		db.Where("user_id = ?", userID).
			Preload("Items").
			Order("updated_at DESC").
			Limit(10).
			Find(&quotes)
		Success(c, quotes)
	}
}
