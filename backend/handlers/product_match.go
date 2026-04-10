package handlers

import (
	"net/http"
	"strings"

	"quotepro-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MatchRequest struct {
	ProductName string `json:"productName"`
	Material    string `json:"material"`
	Size        string `json:"size"`
	Color       string `json:"color"`
	Model       string `json:"model"`
}

func MatchProducts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")

		var req MatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Error(c, http.StatusBadRequest, "参数错误")
			return
		}

		query := db.Where("user_id = ?", userID)

		var conditions []string
		var args []interface{}

		if req.ProductName != "" {
			conditions = append(conditions, "name LIKE ?")
			args = append(args, "%"+req.ProductName+"%")
		}
		if req.Material != "" {
			conditions = append(conditions, "material LIKE ?")
			args = append(args, "%"+req.Material+"%")
		}
		if req.Size != "" {
			conditions = append(conditions, "size LIKE ?")
			args = append(args, "%"+req.Size+"%")
		}
		if req.Color != "" {
			conditions = append(conditions, "color LIKE ?")
			args = append(args, "%"+req.Color+"%")
		}
		if req.Model != "" {
			conditions = append(conditions, "(sku LIKE ? OR name LIKE ?)")
			args = append(args, "%"+req.Model+"%", "%"+req.Model+"%")
		}

		if len(conditions) == 0 {
			Success(c, gin.H{"products": []models.Product{}, "total": 0})
			return
		}

		query = query.Where(strings.Join(conditions, " OR "), args...)

		var products []models.Product
		query.Limit(10).Order("updated_at DESC").Find(&products)

		Success(c, gin.H{
			"products": products,
			"total":    len(products),
		})
	}
}
