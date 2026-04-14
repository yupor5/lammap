package handlers

import (
	"net/http"
	"sort"
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

type scoredProduct struct {
	p     models.Product
	score int
}

func fieldContainsFold(field, needle string) bool {
	needle = strings.TrimSpace(needle)
	if needle == "" {
		return false
	}
	return strings.Contains(strings.ToLower(field), strings.ToLower(needle))
}

// MatchProducts 按用户产品库做推荐：仅对**非空**询盘字段参与匹配与加权，结果按分数排序。
// 权重：名称 5、型号/SKU 4、材质/尺寸 3、颜色 2（与任务文档一致）。
func MatchProducts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")

		var req MatchRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			Error(c, http.StatusBadRequest, "参数错误")
			return
		}

		hasName := strings.TrimSpace(req.ProductName) != ""
		hasMat := strings.TrimSpace(req.Material) != ""
		hasSize := strings.TrimSpace(req.Size) != ""
		hasColor := strings.TrimSpace(req.Color) != ""
		hasModel := strings.TrimSpace(req.Model) != ""

		if !hasName && !hasMat && !hasSize && !hasColor && !hasModel {
			Success(c, gin.H{"products": []models.Product{}, "total": 0})
			return
		}

		q := db.Where("user_id = ?", userID)
		var parts []string
		var args []interface{}
		if hasName {
			parts = append(parts, "name LIKE ?")
			args = append(args, "%"+strings.TrimSpace(req.ProductName)+"%")
		}
		if hasModel {
			parts = append(parts, "(sku LIKE ? OR name LIKE ?)")
			m := strings.TrimSpace(req.Model)
			args = append(args, "%"+m+"%", "%"+m+"%")
		}
		if hasMat {
			parts = append(parts, "material LIKE ?")
			args = append(args, "%"+strings.TrimSpace(req.Material)+"%")
		}
		if hasSize {
			parts = append(parts, "size LIKE ?")
			args = append(args, "%"+strings.TrimSpace(req.Size)+"%")
		}
		if hasColor {
			parts = append(parts, "color LIKE ?")
			args = append(args, "%"+strings.TrimSpace(req.Color)+"%")
		}
		q = q.Where(strings.Join(parts, " OR "), args...)

		var candidates []models.Product
		q.Limit(80).Find(&candidates)

		var scored []scoredProduct
		for _, p := range candidates {
			s := 0
			if hasName && fieldContainsFold(p.Name, req.ProductName) {
				s += 5
			}
			if hasModel && (fieldContainsFold(p.SKU, req.Model) || fieldContainsFold(p.Name, req.Model)) {
				s += 4
			}
			if hasMat && fieldContainsFold(p.Material, req.Material) {
				s += 3
			}
			if hasSize && fieldContainsFold(p.Size, req.Size) {
				s += 3
			}
			if hasColor && fieldContainsFold(p.Color, req.Color) {
				s += 2
			}
			if s > 0 {
				scored = append(scored, scoredProduct{p: p, score: s})
			}
		}

		sort.Slice(scored, func(i, j int) bool {
			if scored[i].score != scored[j].score {
				return scored[i].score > scored[j].score
			}
			return scored[i].p.UpdatedAt.After(scored[j].p.UpdatedAt)
		})

		out := make([]models.Product, 0, 10)
		for _, x := range scored {
			if len(out) >= 10 {
				break
			}
			out = append(out, x.p)
		}

		Success(c, gin.H{
			"products": out,
			"total":    len(out),
		})
	}
}
