package handlers

import (
	"net/http"
	"strings"

	"quotepro-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListTemplates(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		category := c.Query("category")
		lang := strings.TrimSpace(c.Query("language"))

		query := db.Where("user_id = ?", userID)
		if category != "" {
			query = query.Where("category = ?", category)
		}
		if lang != "" {
			query = query.Where("language = ?", lang)
		}

		var templates []models.Template
		query.Order("updated_at DESC").Find(&templates)
		Success(c, templates)
	}
}

func GetTemplate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")
		var tmpl models.Template
		if err := db.First(&tmpl, id).Error; err != nil {
			Error(c, http.StatusNotFound, "模板不存在")
			return
		}
		if tmpl.UserID != userID {
			Error(c, http.StatusNotFound, "模板不存在")
			return
		}
		Success(c, tmpl)
	}
}

func CreateTemplate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		var input struct {
			Name     string `json:"name"`
			Category string `json:"category"`
			Language string `json:"language"`
			Content  string `json:"content"`
			Source   string `json:"source"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
			return
		}
		src := strings.TrimSpace(input.Source)
		if src == "" {
			src = "user"
		}
		if src != "user" && src != "ai" && src != "system" {
			src = "user"
		}
		tmpl := models.Template{
			UserID:    userID,
			Name:      strings.TrimSpace(input.Name),
			Category:  strings.TrimSpace(input.Category),
			Language:  strings.TrimSpace(input.Language),
			Content:   input.Content,
			Source:    src,
		}
		if tmpl.Language == "" {
			tmpl.Language = "zh"
		}
		if err := db.Create(&tmpl).Error; err != nil {
			Error(c, http.StatusInternalServerError, "创建模板失败")
			return
		}
		Success(c, tmpl)
	}
}

func UpdateTemplate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")
		var tmpl models.Template
		if err := db.First(&tmpl, id).Error; err != nil {
			Error(c, http.StatusNotFound, "模板不存在")
			return
		}
		if tmpl.UserID != userID {
			Error(c, http.StatusNotFound, "模板不存在")
			return
		}
		var input struct {
			Name     string `json:"name"`
			Category string `json:"category"`
			Language string `json:"language"`
			Content  string `json:"content"`
			Source   string `json:"source"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			Error(c, http.StatusBadRequest, "参数错误")
			return
		}
		if strings.TrimSpace(input.Source) != "" {
			s := strings.TrimSpace(input.Source)
			if s == "user" || s == "ai" || s == "system" {
				tmpl.Source = s
			}
		}
		if strings.TrimSpace(input.Name) != "" {
			tmpl.Name = input.Name
		}
		if strings.TrimSpace(input.Category) != "" {
			tmpl.Category = input.Category
		}
		if strings.TrimSpace(input.Language) != "" {
			tmpl.Language = input.Language
		}
		tmpl.Content = input.Content
		db.Save(&tmpl)
		Success(c, tmpl)
	}
}

func DeleteTemplate(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")
		var tmpl models.Template
		if err := db.First(&tmpl, id).Error; err != nil {
			Error(c, http.StatusNotFound, "模板不存在")
			return
		}
		if tmpl.UserID != userID {
			Error(c, http.StatusNotFound, "模板不存在")
			return
		}
		if err := db.Delete(&tmpl).Error; err != nil {
			Error(c, http.StatusInternalServerError, "删除失败")
			return
		}
		Success(c, nil)
	}
}
