package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"quotepro-backend/config"
	"quotepro-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const passwordResetTokenTTL = 15 * time.Minute

func randomHexToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

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

// ForgotPassword 为邮箱创建一次性重置令牌（15 分钟有效）。生产环境应发邮件；未配邮件时仅落库。
// 若 config.ExposePasswordResetToken 为 true，响应中会包含 resetToken（仅限本地/内测）。
func ForgotPassword(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
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
			// 防枚举：与存在时返回一致
			Success(c, gin.H{"message": "若该邮箱已注册，请使用邮件或管理员提供的重置令牌完成重置"})
			return
		}

		token, err := randomHexToken(32)
		if err != nil {
			Error(c, http.StatusInternalServerError, "生成令牌失败")
			return
		}

		rec := models.PasswordResetToken{
			Email:     req.Email,
			Token:     token,
			ExpiresAt: time.Now().Add(passwordResetTokenTTL),
			Used:      false,
		}
		if err := db.Create(&rec).Error; err != nil {
			Error(c, http.StatusInternalServerError, "保存重置令牌失败")
			return
		}

		out := gin.H{"message": "若该邮箱已注册，请使用邮件或管理员提供的重置令牌完成重置"}
		if cfg != nil && cfg.ExposePasswordResetToken {
			out["resetToken"] = token
			out["expiresAt"] = rec.ExpiresAt
			out["hint"] = "仅开发环境：EXPOSE_PASSWORD_RESET_TOKEN 已开启，请勿用于生产"
		}
		Success(c, out)
	}
}

// ResetPassword 必须校验 email + token + newPassword。
func ResetPassword(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email       string `json:"email" binding:"required,email"`
			Token       string `json:"token" binding:"required"`
			NewPassword string `json:"newPassword" binding:"required,min=6"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			Error(c, http.StatusBadRequest, "参数错误: 需要 email、token、newPassword")
			return
		}

		var rec models.PasswordResetToken
		if err := db.Where("token = ? AND email = ?", req.Token, req.Email).First(&rec).Error; err != nil {
			Error(c, http.StatusBadRequest, "令牌无效或已过期")
			return
		}
		if rec.Used || time.Now().After(rec.ExpiresAt) {
			Error(c, http.StatusBadRequest, "令牌无效或已过期")
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
		if err := db.Save(&user).Error; err != nil {
			Error(c, http.StatusInternalServerError, "保存失败")
			return
		}

		rec.Used = true
		_ = db.Save(&rec)

		Success(c, gin.H{"message": "密码重置成功"})
	}
}
