package handlers

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"quotepro-backend/config"
	"quotepro-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type zipItem struct {
	name string
	path string
}

// CreateAttachmentZipJob 创建“邮件附件包 zip”任务（仅打包已存在 url 的附件）。
// body:
// { "attachments": [{name,url,selected,source?}] }
func CreateAttachmentZipJob(db *gorm.DB, _ *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		var payload map[string]interface{}
		if err := c.ShouldBindJSON(&payload); err != nil {
			Error(c, http.StatusBadRequest, "参数错误")
			return
		}
		reqBytes, _ := json.Marshal(payload)
		job := models.AttachmentZipJob{
			UserID:      userID,
			Status:      string(models.GenerateJobQueued),
			RequestJSON: string(reqBytes),
		}
		if err := db.Create(&job).Error; err != nil {
			Error(c, http.StatusInternalServerError, "创建任务失败")
			return
		}
		go runAttachmentZipJob(db, job.ID)
		Success(c, gin.H{"jobId": job.ID})
	}
}

func GetAttachmentZipJob(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")
		var job models.AttachmentZipJob
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
			"zipUrl": func() string {
				if result == nil {
					return ""
				}
				if s, ok := result["zipUrl"].(string); ok {
					return s
				}
				return ""
			}(),
			"resultJson": job.ResultJSON,
			"createdAt":  job.CreatedAt,
			"updatedAt":  job.UpdatedAt,
		})
	}
}

func runAttachmentZipJob(db *gorm.DB, jobID uint) {
	var job models.AttachmentZipJob
	if err := db.First(&job, jobID).Error; err != nil {
		return
	}
	db.Model(&models.AttachmentZipJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":     string(models.GenerateJobRunning),
		"updated_at": time.Now(),
	})

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(job.RequestJSON), &payload); err != nil {
		db.Model(&models.AttachmentZipJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
			"status":     string(models.GenerateJobFailed),
			"error_msg":  "参数解析失败",
			"updated_at": time.Now(),
		})
		return
	}

	attachments := toIfaceSlice(payload["attachments"])
	var files []zipItem
	for _, x := range attachments {
		row, ok := x.(map[string]interface{})
		if !ok {
			continue
		}
		sel := true
		if b, ok := row["selected"].(bool); ok {
			sel = b
		}
		if !sel {
			continue
		}
		u := strings.TrimSpace(strVal(row, "url"))
		if u == "" {
			continue
		}
		local, ok := urlToLocalUploadPath(u)
		if !ok {
			continue
		}
		files = append(files, zipItem{name: strings.TrimSpace(strVal(row, "name")), path: local})
	}

	if len(files) == 0 {
		db.Model(&models.AttachmentZipJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
			"status":     string(models.GenerateJobFailed),
			"error_msg":  "没有可打包的附件（需勾选且存在 url）",
			"updated_at": time.Now(),
		})
		return
	}

	dir := filepath.Join("uploads", "generated", "zips", time.Now().Format("2006-01"))
	if err := os.MkdirAll(dir, 0755); err != nil {
		db.Model(&models.AttachmentZipJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
			"status":     string(models.GenerateJobFailed),
			"error_msg":  "创建目录失败",
			"updated_at": time.Now(),
		})
		return
	}

	zipName := fmt.Sprintf("email-attachments-%d.zip", jobID)
	zipPath := filepath.Join(dir, zipName)
	if err := writeZip(zipPath, files); err != nil {
		db.Model(&models.AttachmentZipJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
			"status":     string(models.GenerateJobFailed),
			"error_msg":  "打包失败: " + err.Error(),
			"updated_at": time.Now(),
		})
		return
	}

	result := map[string]interface{}{"zipUrl": "/" + filepath.ToSlash(zipPath)}
	b, _ := json.Marshal(result)
	db.Model(&models.AttachmentZipJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":      string(models.GenerateJobSucceeded),
		"result_json": string(b),
		"updated_at":  time.Now(),
	})
}

func urlToLocalUploadPath(u string) (string, bool) {
	u = strings.TrimSpace(u)
	if u == "" {
		return "", false
	}
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		return "", false
	}
	// 只允许 /uploads/... 或 uploads/...
	u = strings.TrimPrefix(u, "/")
	if !strings.HasPrefix(u, "uploads/") {
		return "", false
	}
	// 防止路径穿越
	clean := filepath.Clean(u)
	if !strings.HasPrefix(filepath.ToSlash(clean), "uploads/") {
		return "", false
	}
	return clean, true
}

func writeZip(zipPath string, files []zipItem) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	defer zw.Close()

	for _, it := range files {
		src, err := os.Open(it.path)
		if err != nil {
			return err
		}
		func() {
			defer src.Close()
			base := safeFileBase(strings.TrimSuffix(it.name, filepath.Ext(it.name)))
			ext := strings.ToLower(filepath.Ext(it.name))
			if ext == "" {
				ext = strings.ToLower(filepath.Ext(it.path))
			}
			if ext == "" {
				ext = ".bin"
			}
			fn := base
			if fn == "" {
				fn = filepath.Base(it.path)
			}
			fn = fn + ext
			w, err2 := zw.Create(fn)
			if err2 != nil {
				err = err2
				return
			}
			_, err = io.Copy(w, src)
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

