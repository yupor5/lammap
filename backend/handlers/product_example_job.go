package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"quotepro-backend/config"
	"quotepro-backend/models"
	"quotepro-backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createProductExampleJobReq struct {
	Count     int    `json:"count"`
	ExtraHint string `json:"extraHint"`
}

// CreateProductExampleJob 创建产品示例生成任务，立即返回 jobId，前端轮询 GET /products/ai-example-jobs/:id。
func CreateProductExampleJob(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")

		var body createProductExampleJobReq
		if err := c.ShouldBindJSON(&body); err != nil {
			Error(c, http.StatusBadRequest, "参数错误")
			return
		}
		if body.Count <= 0 {
			body.Count = 3
		}
		if body.Count > 5 {
			body.Count = 5
		}
		body.ExtraHint = strings.TrimSpace(body.ExtraHint)

		reqBytes, _ := json.Marshal(body)
		job := models.ProductExampleJob{
			UserID:      userID,
			Status:      string(models.GenerateJobQueued),
			RequestJSON: string(reqBytes),
		}
		if err := db.Create(&job).Error; err != nil {
			Error(c, http.StatusInternalServerError, "创建任务失败")
			return
		}

		go runProductExampleJob(db, cfg, job.ID)

		Success(c, gin.H{"jobId": job.ID})
	}
}

// GetProductExampleJob 查询任务；成功时返回 items 数组（可直接用于 /products/import）。
func GetProductExampleJob(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")

		var job models.ProductExampleJob
		if err := db.First(&job, id).Error; err != nil {
			Error(c, http.StatusNotFound, "任务不存在")
			return
		}
		if job.UserID != userID {
			Error(c, http.StatusNotFound, "任务不存在")
			return
		}

		var result interface{} = nil
		if strings.TrimSpace(job.ResultJSON) != "" {
			var m map[string]interface{}
			if json.Unmarshal([]byte(job.ResultJSON), &m) == nil {
				result = m
			}
		}

		Success(c, gin.H{
			"id":         job.ID,
			"status":     job.Status,
			"errorMsg":   job.ErrorMsg,
			"result":     result,
			"resultJson": job.ResultJSON,
			"createdAt":  job.CreatedAt,
			"updatedAt":  job.UpdatedAt,
		})
	}
}

func runProductExampleJob(db *gorm.DB, cfg *config.Config, jobID uint) {
	var job models.ProductExampleJob
	if err := db.First(&job, jobID).Error; err != nil {
		return
	}

	db.Model(&models.ProductExampleJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":     string(models.GenerateJobRunning),
		"updated_at": time.Now(),
	})

	var body createProductExampleJobReq
	if err := json.Unmarshal([]byte(job.RequestJSON), &body); err != nil {
		db.Model(&models.ProductExampleJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
			"status":     string(models.GenerateJobFailed),
			"error_msg":  "参数解析失败",
			"updated_at": time.Now(),
		})
		return
	}

	count := body.Count
	if count < 1 {
		count = 3
	}
	if count > 5 {
		count = 5
	}

	items, err := services.GenerateProductExamplesWithAI(cfg, count, body.ExtraHint)
	if err != nil {
		db.Model(&models.ProductExampleJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
			"status":     string(models.GenerateJobFailed),
			"error_msg":  "AI 生成失败: " + err.Error(),
			"updated_at": time.Now(),
		})
		return
	}
	if len(items) == 0 {
		db.Model(&models.ProductExampleJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
			"status":     string(models.GenerateJobFailed),
			"error_msg":  "AI 未生成有效示例",
			"updated_at": time.Now(),
		})
		return
	}

	b, _ := json.Marshal(map[string]interface{}{"items": items})
	db.Model(&models.ProductExampleJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":      string(models.GenerateJobSucceeded),
		"result_json": string(b),
		"updated_at":  time.Now(),
	})
}

