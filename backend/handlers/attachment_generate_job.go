package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"quotepro-backend/config"
	"quotepro-backend/models"
	"quotepro-backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateAttachmentGenerateJob 创建“AI 生成单个附件”任务。
// 期望 body:
// {
//   "params": {...},
//   "quote": {...},
//   "attachment": {"name": "...", "selected": true, "url": "", "source": "ai"}
// }
func CreateAttachmentGenerateJob(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		var payload map[string]interface{}
		if err := c.ShouldBindJSON(&payload); err != nil {
			Error(c, http.StatusBadRequest, "参数错误")
			return
		}

		att, _ := payload["attachment"].(map[string]interface{})
		if att == nil {
			Error(c, http.StatusBadRequest, "缺少 attachment")
			return
		}
		name := strings.TrimSpace(strVal(att, "name"))
		if name == "" {
			Error(c, http.StatusBadRequest, "附件名称为空")
			return
		}
		if strings.TrimSpace(strVal(att, "url")) != "" {
			Error(c, http.StatusBadRequest, "附件已存在 url，无需生成")
			return
		}
		if src := strings.TrimSpace(strVal(att, "source")); src == "upload" {
			Error(c, http.StatusBadRequest, "上传附件不支持 AI 生成")
			return
		}

		reqBytes, _ := json.Marshal(payload)
		job := models.AttachmentGenerateJob{
			UserID:      userID,
			Status:      string(models.GenerateJobQueued),
			RequestJSON: string(reqBytes),
		}
		if err := db.Create(&job).Error; err != nil {
			Error(c, http.StatusInternalServerError, "创建任务失败")
			return
		}

		go runAttachmentGenerateJob(db, cfg, job.ID)
		Success(c, gin.H{"jobId": job.ID})
	}
}

func GetAttachmentGenerateJob(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")

		var job models.AttachmentGenerateJob
		if err := db.First(&job, id).Error; err != nil {
			Error(c, http.StatusNotFound, "任务不存在")
			return
		}
		if job.UserID != userID {
			Error(c, http.StatusNotFound, "任务不存在")
			return
		}

		var result map[string]interface{}
		if strings.TrimSpace(job.ResultJSON) != "" {
			_ = json.Unmarshal([]byte(job.ResultJSON), &result)
		}

		Success(c, gin.H{
			"id":       job.ID,
			"status":   job.Status,
			"errorMsg": job.ErrorMsg,
			"attachment": func() interface{} {
				if result == nil {
					return nil
				}
				return result["attachment"]
			}(),
			"resultJson": job.ResultJSON,
			"createdAt":  job.CreatedAt,
			"updatedAt":  job.UpdatedAt,
		})
	}
}

func runAttachmentGenerateJob(db *gorm.DB, cfg *config.Config, jobID uint) {
	var job models.AttachmentGenerateJob
	if err := db.First(&job, jobID).Error; err != nil {
		return
	}
	db.Model(&models.AttachmentGenerateJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":     string(models.GenerateJobRunning),
		"updated_at": time.Now(),
	})

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(job.RequestJSON), &payload); err != nil {
		db.Model(&models.AttachmentGenerateJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
			"status":     string(models.GenerateJobFailed),
			"error_msg":  "参数解析失败",
			"updated_at": time.Now(),
		})
		return
	}

	params, _ := payload["params"].(map[string]interface{})
	quote, _ := payload["quote"].(map[string]interface{})
	att, _ := payload["attachment"].(map[string]interface{})
	name := strings.TrimSpace(strVal(att, "name"))

	dir := filepath.Join("uploads", "generated", "attachments", time.Now().Format("2006-01"), fmt.Sprintf("attjob-%d", jobID))
	if err := os.MkdirAll(dir, 0755); err != nil {
		db.Model(&models.AttachmentGenerateJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
			"status":     string(models.GenerateJobFailed),
			"error_msg":  "创建目录失败",
			"updated_at": time.Now(),
		})
		return
	}

	base := safeFileBase(strings.TrimSuffix(name, filepath.Ext(name)))
	if base == "" {
		base = "attachment"
	}
	dst := filepath.Join(dir, base+".html")
	html, err := services.GenerateAttachmentHTML(cfg, name, params, quote)
	if err != nil || strings.TrimSpace(html) == "" {
		html = fmt.Sprintf("<!doctype html><html><head><meta charset=\"utf-8\"><title>%s</title></head><body><h3>%s</h3><p>生成失败，请稍后重试。</p></body></html>", name, name)
	}
	_ = os.WriteFile(dst, []byte(html), 0644)

	result := map[string]interface{}{
		"attachment": map[string]interface{}{
			"name": name,
			"url":  "/" + filepath.ToSlash(dst),
		},
	}
	b, _ := json.Marshal(result)
	db.Model(&models.AttachmentGenerateJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":      string(models.GenerateJobSucceeded),
		"result_json": string(b),
		"updated_at":  time.Now(),
	})
}

