package handlers

import (
	"net/http"

	"quotepro-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ChangePassword(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")

		var req struct {
			OldPassword string `json:"oldPassword" binding:"required"`
			NewPassword string `json:"newPassword" binding:"required,min=6"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
			return
		}

		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			Error(c, http.StatusNotFound, "用户不存在")
			return
		}

		if !user.CheckPassword(req.OldPassword) {
			Error(c, http.StatusBadRequest, "原密码错误")
			return
		}

		if err := user.SetPassword(req.NewPassword); err != nil {
			Error(c, http.StatusInternalServerError, "密码加密失败")
			return
		}

		db.Save(&user)
		Success(c, gin.H{"message": "密码修改成功"})
	}
}

// ForgotPassword - MVP 版: 仅验证邮箱是否存在并返回确认
// ⚠️ 需要用户提供: 完整版需接入 SMTP 邮件服务发送重置链接
func ForgotPassword(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email" binding:"required,email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			Error(c, http.StatusBadRequest, "请输入有效邮箱")
			return
		}

		var user models.User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			Success(c, gin.H{"message": "如果邮箱存在，重置邮件已发送"})
			return
		}

		// TODO: 接入 SMTP 服务发送重置链接
		// 目前 MVP 阶段直接返回成功提示，不实际发送邮件
		Success(c, gin.H{"message": "如果邮箱存在，重置邮件已发送"})
	}
}

func ResetPassword(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email       string `json:"email" binding:"required,email"`
			NewPassword string `json:"newPassword" binding:"required,min=6"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
			return
		}

		var user models.User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			Error(c, http.StatusNotFound, "用户不存在")
			return
		}

		if err := user.SetPassword(req.NewPassword); err != nil {
			Error(c, http.StatusInternalServerError, "密码加密失败")
			return
		}

		db.Save(&user)
		Success(c, gin.H{"message": "密码重置成功"})
	}
}
