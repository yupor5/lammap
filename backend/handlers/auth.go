package handlers

import (
	"net/http"
	"time"

	"quotepro-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Company  string `json:"company"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func Register(db *gorm.DB, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
			return
		}

		var existing models.User
		if db.Where("email = ?", req.Email).First(&existing).Error == nil {
			Error(c, http.StatusConflict, "邮箱已注册")
			return
		}

		user := models.User{
			Email:   req.Email,
			Name:    req.Name,
			Company: req.Company,
		}
		if err := user.SetPassword(req.Password); err != nil {
			Error(c, http.StatusInternalServerError, "密码加密失败")
			return
		}

		if err := db.Create(&user).Error; err != nil {
			Error(c, http.StatusInternalServerError, "创建用户失败")
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": user.ID,
			"email":   user.Email,
			"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		})

		tokenStr, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			Error(c, http.StatusInternalServerError, "生成 token 失败")
			return
		}

		Success(c, gin.H{
			"token": tokenStr,
			"user":  user,
		})
	}
}

func Login(db *gorm.DB, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
			return
		}

		var user models.User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			Error(c, http.StatusUnauthorized, "邮箱或密码错误")
			return
		}

		if !user.CheckPassword(req.Password) {
			Error(c, http.StatusUnauthorized, "邮箱或密码错误")
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": user.ID,
			"email":   user.Email,
			"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
		})

		tokenStr, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			Error(c, http.StatusInternalServerError, "生成 token 失败")
			return
		}

		Success(c, gin.H{
			"token": tokenStr,
			"user":  user,
		})
	}
}

func Profile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			Error(c, http.StatusNotFound, "用户不存在")
			return
		}
		Success(c, user)
	}
}
