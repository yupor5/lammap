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

// CreateAttachmentPackJob 创建“邮件附件包”生成任务。
// 期望 body:
// {
//   "params": {...},          // ParsedParams
//   "quote": {...},           // Quote（至少包含 replyVersions/language 等）
//   "attachments": [{name, selected, url?}]
// }
func CreateAttachmentPackJob(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")

		var payload map[string]interface{}
		if err := c.ShouldBindJSON(&payload); err != nil {
			Error(c, http.StatusBadRequest, "参数错误")
			return
		}
		reqBytes, _ := json.Marshal(payload)

		job := models.AttachmentPackJob{
			UserID:      userID,
			Status:      string(models.GenerateJobQueued),
			RequestJSON: string(reqBytes),
		}
		if err := db.Create(&job).Error; err != nil {
			Error(c, http.StatusInternalServerError, "创建任务失败")
			return
		}

		go runAttachmentPackJob(db, cfg, job.ID)
		Success(c, gin.H{"jobId": job.ID})
	}
}

// GetAttachmentPackJob 获取任务状态；成功时返回 attachments（带 url）。
func GetAttachmentPackJob(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")

		var job models.AttachmentPackJob
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
			"id":         job.ID,
			"status":     job.Status,
			"errorMsg":   job.ErrorMsg,
			"attachments": func() interface{} {
				if result == nil {
					return nil
				}
				return result["attachments"]
			}(),
			"resultJson": job.ResultJSON,
			"createdAt":  job.CreatedAt,
			"updatedAt":  job.UpdatedAt,
		})
	}
}

func runAttachmentPackJob(db *gorm.DB, cfg *config.Config, jobID uint) {
	var job models.AttachmentPackJob
	if err := db.First(&job, jobID).Error; err != nil {
		return
	}

	db.Model(&models.AttachmentPackJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":     string(models.GenerateJobRunning),
		"updated_at": time.Now(),
	})

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(job.RequestJSON), &payload); err != nil {
		db.Model(&models.AttachmentPackJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
			"status":     string(models.GenerateJobFailed),
			"error_msg":  "参数解析失败",
			"updated_at": time.Now(),
		})
		return
	}

	params, _ := payload["params"].(map[string]interface{})
	quote, _ := payload["quote"].(map[string]interface{})
	attachmentsAny := payload["attachments"]
	attachments := toIfaceSlice(attachmentsAny)

	// 生成目录：uploads/generated/attachments/<YYYY-MM>/job-<id>/
	dir := filepath.Join("uploads", "generated", "attachments", time.Now().Format("2006-01"), fmt.Sprintf("job-%d", jobID))
	if err := os.MkdirAll(dir, 0755); err != nil {
		db.Model(&models.AttachmentPackJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
			"status":     string(models.GenerateJobFailed),
			"error_msg":  "创建目录失败",
			"updated_at": time.Now(),
		})
		return
	}

	var out []map[string]interface{}
	for i, x := range attachments {
		row, ok := x.(map[string]interface{})
		if !ok {
			continue
		}
		name := strings.TrimSpace(strVal(row, "name"))
		if name == "" {
			continue
		}
		sel := true
		if b, ok := row["selected"].(bool); ok {
			sel = b
		}
		// 若已有 url 且可访问，则直接返回
		u := strings.TrimSpace(strVal(row, "url"))
		if u != "" {
			out = append(out, map[string]interface{}{
				"name":     name,
				"url":      normalizeUploadURL(u),
				"selected": sel,
			})
			continue
		}

		base := safeFileBase(strings.TrimSuffix(name, filepath.Ext(name)))
		if base == "" {
			base = fmt.Sprintf("attachment-%d", i+1)
		}
		filename := base + ".html"
		dst := filepath.Join(dir, filename)

		// quote 作为上下文（含 replyVersions.language）
		html, err := services.GenerateAttachmentHTML(cfg, name, params, quote)
		if err != nil || strings.TrimSpace(html) == "" {
			html = fmt.Sprintf("<!doctype html><html><head><meta charset=\"utf-8\"><title>%s</title></head><body><h3>%s</h3><p>生成失败，请稍后重试。</p></body></html>", name, name)
		}
		_ = os.WriteFile(dst, []byte(html), 0644)

		out = append(out, map[string]interface{}{
			"name":     name,
			"url":      "/" + filepath.ToSlash(dst),
			"selected": sel,
		})
	}

	b, _ := json.Marshal(map[string]interface{}{"attachments": out})
	db.Model(&models.AttachmentPackJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":      string(models.GenerateJobSucceeded),
		"result_json": string(b),
		"updated_at":  time.Now(),
	})
}

func normalizeUploadURL(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return ""
	}
	if strings.HasPrefix(u, "/") {
		return u
	}
	// 兼容存的是 uploads/xxx
	if strings.HasPrefix(u, "uploads/") {
		return "/" + u
	}
	return u
}

