package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"quotepro-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListProducts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		search := c.Query("search")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

		var products []models.Product
		query := db.Where("user_id = ?", userID)

		if search != "" {
			query = query.Where("name LIKE ? OR sku LIKE ? OR category LIKE ?",
				"%"+search+"%", "%"+search+"%", "%"+search+"%")
		}

		var total int64
		query.Model(&models.Product{}).Count(&total)

		query.Offset((page - 1) * pageSize).Limit(pageSize).Order("updated_at DESC").Find(&products)

		Success(c, gin.H{
			"items": products,
			"total": total,
			"page":  page,
		})
	}
}

func GetProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")
		var product models.Product
		if err := db.First(&product, id).Error; err != nil {
			Error(c, http.StatusNotFound, "产品不存在")
			return
		}
		if product.UserID != userID {
			Error(c, http.StatusNotFound, "产品不存在")
			return
		}
		Success(c, product)
	}
}

func CreateProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		var product models.Product
		if err := c.ShouldBindJSON(&product); err != nil {
			Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
			return
		}
		product.UserID = userID
		if err := db.Create(&product).Error; err != nil {
			Error(c, http.StatusInternalServerError, "创建产品失败")
			return
		}
		Success(c, product)
	}
}

func UpdateProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")
		var product models.Product
		if err := db.First(&product, id).Error; err != nil {
			Error(c, http.StatusNotFound, "产品不存在")
			return
		}
		if product.UserID != userID {
			Error(c, http.StatusNotFound, "产品不存在")
			return
		}
		if err := c.ShouldBindJSON(&product); err != nil {
			Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
			return
		}
		product.UserID = userID
		db.Save(&product)
		Success(c, product)
	}
}

func DeleteProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")
		var product models.Product
		if err := db.First(&product, id).Error; err != nil {
			Error(c, http.StatusNotFound, "产品不存在")
			return
		}
		if product.UserID != userID {
			Error(c, http.StatusNotFound, "产品不存在")
			return
		}
		if err := db.Delete(&product).Error; err != nil {
			Error(c, http.StatusInternalServerError, "删除失败")
			return
		}
		Success(c, nil)
	}
}

func ImportProducts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")

		var products []models.Product
		if err := c.ShouldBindJSON(&products); err != nil {
			Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
			return
		}

		if len(products) == 0 {
			Error(c, http.StatusBadRequest, "产品列表为空")
			return
		}

		created := 0
		for i := range products {
			products[i].UserID = userID
			products[i].ID = 0
			if err := db.Create(&products[i]).Error; err == nil {
				created++
			}
		}

		Success(c, gin.H{
			"total":   len(products),
			"created": created,
			"message": fmt.Sprintf("成功导入 %d/%d 个产品", created, len(products)),
		})
	}
}
