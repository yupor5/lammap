package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"quotepro-backend/config"
	"quotepro-backend/models"
	"quotepro-backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateGenerateJob 创建异步生成任务，立即返回 jobId。
func CreateGenerateJob(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")

		var params map[string]interface{}
		if err := c.ShouldBindJSON(&params); err != nil {
			Error(c, http.StatusBadRequest, "参数错误")
			return
		}

		reqBytes, _ := json.Marshal(params)
		job := models.GenerateJob{
			UserID:      userID,
			Status:      string(models.GenerateJobQueued),
			RequestJSON: string(reqBytes),
		}
		if err := db.Create(&job).Error; err != nil {
			Error(c, http.StatusInternalServerError, "创建任务失败")
			return
		}

		// 后台执行
		go runGenerateJob(db, cfg, job.ID)

		Success(c, gin.H{"jobId": job.ID})
	}
}

// GetGenerateJob 查询任务状态；成功时返回 result（已 normalize），失败时返回 errorMsg。
func GetGenerateJob(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")

		var job models.GenerateJob
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
			"id":        job.ID,
			"status":    job.Status,
			"errorMsg":  job.ErrorMsg,
			"result":    result,
			"resultJson": func() string {
				// 前端兜底解析用；同时方便排查“解析失败导致 result 为 null”的情况
				if strings.TrimSpace(job.ResultJSON) == "" {
					return ""
				}
				return job.ResultJSON
			}(),
			"createdAt": job.CreatedAt,
			"updatedAt": job.UpdatedAt,
		})
	}
}

func runGenerateJob(db *gorm.DB, cfg *config.Config, jobID uint) {
	// 读取任务
	var job models.GenerateJob
	if err := db.First(&job, jobID).Error; err != nil {
		return
	}

	db.Model(&models.GenerateJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":     string(models.GenerateJobRunning),
		"updated_at": time.Now(),
	})

	var params map[string]interface{}
	if err := json.Unmarshal([]byte(job.RequestJSON), &params); err != nil {
		db.Model(&models.GenerateJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
			"status":     string(models.GenerateJobFailed),
			"error_msg":  "参数解析失败",
			"updated_at": time.Now(),
		})
		return
	}

	clean, hints := splitQuoteMeta(params)

	raw, err := services.GenerateQuoteWithAI(cfg, clean, hints)
	if err != nil {
		db.Model(&models.GenerateJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
			"status":     string(models.GenerateJobFailed),
			"error_msg":  "报价生成失败: " + err.Error(),
			"updated_at": time.Now(),
		})
		return
	}

	normalized := normalizeGenerateResponse(raw, strVal(clean, "customerName"), strVal(clean, "currency"))
	if arr, ok := normalized["items"].([]map[string]interface{}); ok && len(arr) == 0 {
		normalized["items"] = []map[string]interface{}{{
			"productName": strVal(clean, "productName"),
			"model":       strVal(clean, "model"),
			"specs":       "",
			"quantity":    intVal(clean["quantity"]),
			"unitPrice":   0,
			"totalPrice":  0,
			"remark":      "",
		}}
	}

	b, _ := json.Marshal(normalized)
	db.Model(&models.GenerateJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":      string(models.GenerateJobSucceeded),
		"result_json": string(b),
		"updated_at":  time.Now(),
	})

	// 打印生成结果（截断避免日志过大）
	s := string(b)
	if len(s) > 4000 {
		s = s[:4000] + "...(truncated)"
	}
	log.Printf("[generate-job] id=%d status=succeeded result=%s", jobID, s)
}

