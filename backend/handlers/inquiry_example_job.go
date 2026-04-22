package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"quotepro-backend/config"
	"quotepro-backend/models"
	"quotepro-backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateInquiryExampleJob 创建询盘示例批量生成任务，立即返回 jobId，前端轮询 GET /ai/inquiry-example-jobs/:id。
func CreateInquiryExampleJob(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")

		var body struct {
			GroupName   string `json:"groupName"`
			GroupPrompt string `json:"groupPrompt"`
			Language    string `json:"language"`
			Count       int    `json:"count"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			Error(c, http.StatusBadRequest, "参数错误")
			return
		}
		name := strings.TrimSpace(body.GroupName)
		if name == "" {
			Error(c, http.StatusBadRequest, "请填写分组名称")
			return
		}
		prompt := strings.TrimSpace(body.GroupPrompt)
		if prompt == "" {
			Error(c, http.StatusBadRequest, "请填写分组提示词")
			return
		}

		reqBytes, _ := json.Marshal(body)
		job := models.InquiryExampleJob{
			UserID:      userID,
			Status:      string(models.GenerateJobQueued),
			RequestJSON: string(reqBytes),
		}
		if err := db.Create(&job).Error; err != nil {
			Error(c, http.StatusInternalServerError, "创建任务失败")
			return
		}

		go runInquiryExampleJob(db, cfg, job.ID)

		Success(c, gin.H{"jobId": job.ID})
	}
}

// GetInquiryExampleJob 查询任务；成功时 result 中为 examples 数组。
func GetInquiryExampleJob(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")

		var job models.InquiryExampleJob
		if err := db.First(&job, id).Error; err != nil {
			Error(c, http.StatusNotFound, "任务不存在")
			return
		}
		if job.UserID != userID {
			Error(c, http.StatusNotFound, "任务不存在")
			return
		}

		var examples interface{}
		if strings.TrimSpace(job.ResultJSON) != "" {
			var m map[string]interface{}
			if json.Unmarshal([]byte(job.ResultJSON), &m) == nil {
				examples = m["examples"]
			}
		}

		Success(c, gin.H{
			"id":         job.ID,
			"status":     job.Status,
			"errorMsg":   job.ErrorMsg,
			"examples":   examples,
			"resultJson": job.ResultJSON,
			"createdAt":  job.CreatedAt,
			"updatedAt":  job.UpdatedAt,
		})
	}
}

func runInquiryExampleJob(db *gorm.DB, cfg *config.Config, jobID uint) {
	var job models.InquiryExampleJob
	if err := db.First(&job, jobID).Error; err != nil {
		return
	}

	db.Model(&models.InquiryExampleJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":     string(models.GenerateJobRunning),
		"updated_at": time.Now(),
	})

	var body struct {
		GroupName   string `json:"groupName"`
		GroupPrompt string `json:"groupPrompt"`
		Language    string `json:"language"`
		Count       int    `json:"count"`
	}
	if err := json.Unmarshal([]byte(job.RequestJSON), &body); err != nil {
		db.Model(&models.InquiryExampleJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
			"status":     string(models.GenerateJobFailed),
			"error_msg":  "参数解析失败",
			"updated_at": time.Now(),
		})
		return
	}

	lang := strings.ToLower(strings.TrimSpace(body.Language))
	if lang != "zh" && lang != "en" {
		lang = "en"
	}
	count := body.Count
	if count < 1 {
		count = 3
	}
	if count > 10 {
		count = 10
	}

	rows, err := services.GenerateInquiryExamplesWithAI(cfg, body.GroupName, body.GroupPrompt, lang, count)
	if err != nil {
		db.Model(&models.InquiryExampleJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
			"status":     string(models.GenerateJobFailed),
			"error_msg":  "AI 生成失败: " + err.Error(),
			"updated_at": time.Now(),
		})
		return
	}

	var out []map[string]interface{}
	for _, row := range rows {
		title := strings.TrimSpace(fmt.Sprint(row["title"]))
		l := strings.ToLower(strings.TrimSpace(fmt.Sprint(row["lang"])))
		if l != "zh" && l != "en" {
			l = lang
		}
		content := strings.TrimSpace(fmt.Sprint(row["content"]))
		if content == "" {
			continue
		}
		out = append(out, map[string]interface{}{
			"title":   title,
			"lang":    l,
			"content": content,
		})
	}
	if len(out) == 0 {
		db.Model(&models.InquiryExampleJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
			"status":     string(models.GenerateJobFailed),
			"error_msg":  "AI 未生成有效示例",
			"updated_at": time.Now(),
		})
		return
	}

	b, _ := json.Marshal(map[string]interface{}{"examples": out})
	db.Model(&models.InquiryExampleJob{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":      string(models.GenerateJobSucceeded),
		"result_json": string(b),
		"updated_at":  time.Now(),
	})
}
