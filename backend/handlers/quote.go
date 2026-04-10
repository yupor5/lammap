package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"quotepro-backend/config"
	"quotepro-backend/models"
	"quotepro-backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ParseRequirement(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		content := strings.TrimSpace(c.PostForm("content"))
		if content == "" && strings.Contains(strings.ToLower(c.GetHeader("Content-Type")), "application/json") {
			var body struct {
				Content string `json:"content"`
			}
			if err := c.ShouldBindJSON(&body); err == nil {
				content = strings.TrimSpace(body.Content)
			}
		}
		if content == "" {
			Error(c, http.StatusBadRequest, "请输入客户需求内容")
			return
		}

		result, err := services.ParseRequirementWithAI(cfg, content)
		if err != nil {
			Error(c, http.StatusInternalServerError, "AI 解析失败: "+err.Error())
			return
		}
		Success(c, normalizeParsedParams(result))
	}
}

func GenerateQuote(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var params map[string]interface{}
		if err := c.ShouldBindJSON(&params); err != nil {
			Error(c, http.StatusBadRequest, "参数错误")
			return
		}

		clean, hints := splitQuoteMeta(params)
		replyLang := ""
		if v, ok := hints["_replyLanguage"].(string); ok {
			replyLang = v
		}

		result, err := services.GenerateQuoteWithAI(cfg, clean, hints)
		if err != nil {
			Error(c, http.StatusInternalServerError, "报价生成失败: "+err.Error())
			return
		}
		normalized := normalizeGenerateResponse(result, strVal(clean, "customerName"), strVal(clean, "currency"), replyLang)
		if arr, ok := normalized["items"].([]map[string]interface{}); ok && len(arr) == 0 {
			normalized["items"] = []map[string]interface{}{{
				"productName": strVal(clean, "productName"),
				"model":       strVal(clean, "model"),
				"specs":       "",
				"quantity":    intVal(clean["quantity"]),
				"unitPrice":   0,
				"totalPrice":  0,
				"remark":      "",
			}}
		}
		Success(c, normalized)
	}
}

func CreateQuote(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")

		var input struct {
			CustomerName       string             `json:"customerName"`
			Country            string             `json:"country"`
			Currency           string             `json:"currency"`
			DeliveryAddress    string             `json:"deliveryAddress"`
			Status             string             `json:"status"`
			TotalAmount        float64            `json:"totalAmount"`
			LeadTime           string             `json:"leadTime"`
			Remarks            string             `json:"remarks"`
			Terms              string             `json:"terms"`
			RawRequirement     string             `json:"rawRequirement"`
			ParsedParams       interface{}        `json:"parsedParams"`
			ReplyVersions      interface{}        `json:"replyVersions"`
			ConfirmationList   interface{}        `json:"confirmationList"`
			AttachmentList     interface{}        `json:"attachmentList"`
			TemplateMeta       interface{}        `json:"templateMeta"`
			RenderedContents   interface{}        `json:"renderedContents"`
			Items              []models.QuoteItem `json:"items"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			Error(c, http.StatusBadRequest, "参数错误: "+err.Error())
			return
		}

		parsedJSON, _ := json.Marshal(input.ParsedParams)
		replyJSON, _ := json.Marshal(input.ReplyVersions)
		confirmJSON, _ := json.Marshal(input.ConfirmationList)
		attachJSON, _ := json.Marshal(input.AttachmentList)
		tmplMetaJSON, _ := json.Marshal(input.TemplateMeta)
		renderedJSON, _ := json.Marshal(input.RenderedContents)

		total := input.TotalAmount
		if total == 0 && len(input.Items) > 0 {
			for _, it := range input.Items {
				total += it.TotalPrice
			}
		}

		quote := models.Quote{
			UserID:           userID,
			QuoteNumber:      fmt.Sprintf("QT-%s-%03d", time.Now().Format("2006"), time.Now().UnixNano()%1000),
			CustomerName:     input.CustomerName,
			Country:          input.Country,
			Currency:         input.Currency,
			DeliveryAddress:  input.DeliveryAddress,
			Status:           input.Status,
			TotalAmount:      total,
			LeadTime:         input.LeadTime,
			Remarks:          input.Remarks,
			Terms:            input.Terms,
			RawRequirement:   input.RawRequirement,
			ParsedParams:     string(parsedJSON),
			ReplyVersions:    string(replyJSON),
			ConfirmationList: string(confirmJSON),
			AttachmentList:   string(attachJSON),
			TemplateMeta:     string(tmplMetaJSON),
			RenderedContents: string(renderedJSON),
			Items:            input.Items,
		}

		if quote.Status == "" {
			quote.Status = "草稿"
		}

		if err := db.Create(&quote).Error; err != nil {
			Error(c, http.StatusInternalServerError, "保存报价失败")
			return
		}
		Success(c, quote)
	}
}

func ListQuotes(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
		customer := c.Query("customer")
		status := c.Query("status")

		query := db.Where("user_id = ?", userID)
		if customer != "" {
			query = query.Where("customer_name LIKE ?", "%"+customer+"%")
		}
		if status != "" {
			query = query.Where("status = ?", status)
		}

		var total int64
		query.Model(&models.Quote{}).Count(&total)

		var quotes []models.Quote
		query.Preload("Items").
			Offset((page-1)*pageSize).Limit(pageSize).
			Order("updated_at DESC").
			Find(&quotes)

		Success(c, gin.H{
			"items": quotes,
			"total": total,
			"page":  page,
		})
	}
}

func GetQuote(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")
		var quote models.Quote
		if err := db.Preload("Items").First(&quote, id).Error; err != nil {
			Error(c, http.StatusNotFound, "报价不存在")
			return
		}
		if quote.UserID != userID {
			Error(c, http.StatusNotFound, "报价不存在")
			return
		}
		Success(c, quote)
	}
}

func UpdateQuote(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")
		var quote models.Quote
		if err := db.First(&quote, id).Error; err != nil {
			Error(c, http.StatusNotFound, "报价不存在")
			return
		}
		if quote.UserID != userID {
			Error(c, http.StatusNotFound, "报价不存在")
			return
		}

		var input struct {
			CustomerName     string             `json:"customerName"`
			Country          string             `json:"country"`
			Currency         string             `json:"currency"`
			DeliveryAddress  string             `json:"deliveryAddress"`
			Status           string             `json:"status"`
			TotalAmount      float64            `json:"totalAmount"`
			LeadTime         string             `json:"leadTime"`
			Remarks          string             `json:"remarks"`
			Terms            string             `json:"terms"`
			RawRequirement   string             `json:"rawRequirement"`
			ParsedParams     interface{}        `json:"parsedParams"`
			ReplyVersions    interface{}        `json:"replyVersions"`
			ConfirmationList interface{}        `json:"confirmationList"`
			AttachmentList   interface{}        `json:"attachmentList"`
			TemplateMeta     interface{}        `json:"templateMeta"`
			RenderedContents interface{}        `json:"renderedContents"`
			Items            []models.QuoteItem `json:"items"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			Error(c, http.StatusBadRequest, "参数错误")
			return
		}

		quote.CustomerName = input.CustomerName
		quote.Country = input.Country
		quote.Currency = input.Currency
		quote.DeliveryAddress = input.DeliveryAddress
		if input.Status != "" {
			quote.Status = input.Status
		}
		quote.TotalAmount = input.TotalAmount
		quote.LeadTime = input.LeadTime
		quote.Remarks = input.Remarks
		quote.Terms = input.Terms
		quote.RawRequirement = input.RawRequirement

		if input.ParsedParams != nil {
			b, _ := json.Marshal(input.ParsedParams)
			quote.ParsedParams = string(b)
		}
		if input.ReplyVersions != nil {
			b, _ := json.Marshal(input.ReplyVersions)
			quote.ReplyVersions = string(b)
		}
		if input.ConfirmationList != nil {
			b, _ := json.Marshal(input.ConfirmationList)
			quote.ConfirmationList = string(b)
		}
		if input.AttachmentList != nil {
			b, _ := json.Marshal(input.AttachmentList)
			quote.AttachmentList = string(b)
		}
		if input.TemplateMeta != nil {
			b, _ := json.Marshal(input.TemplateMeta)
			quote.TemplateMeta = string(b)
		}
		if input.RenderedContents != nil {
			b, _ := json.Marshal(input.RenderedContents)
			quote.RenderedContents = string(b)
		}

		if err := db.Save(&quote).Error; err != nil {
			Error(c, http.StatusInternalServerError, "保存失败")
			return
		}

		if input.Items != nil {
			db.Where("quote_id = ?", quote.ID).Delete(&models.QuoteItem{})
			for i := range input.Items {
				input.Items[i].ID = 0
				input.Items[i].QuoteID = quote.ID
			}
			if len(input.Items) > 0 {
				db.Create(&input.Items)
			}
			var sum float64
			for _, it := range input.Items {
				sum += it.TotalPrice
			}
			if sum > 0 {
				quote.TotalAmount = sum
				db.Model(&quote).Update("total_amount", sum)
			}
		}

		var out models.Quote
		db.Preload("Items").First(&out, quote.ID)
		Success(c, out)
	}
}

func DeleteQuote(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")
		var quote models.Quote
		if err := db.First(&quote, id).Error; err != nil {
			Error(c, http.StatusNotFound, "报价不存在")
			return
		}
		if quote.UserID != userID {
			Error(c, http.StatusNotFound, "报价不存在")
			return
		}
		db.Where("quote_id = ?", id).Delete(&models.QuoteItem{})
		if err := db.Delete(&models.Quote{}, id).Error; err != nil {
			Error(c, http.StatusInternalServerError, "删除失败")
			return
		}
		Success(c, nil)
	}
}

func DuplicateQuote(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		id := c.Param("id")
		var original models.Quote
		if err := db.Preload("Items").First(&original, id).Error; err != nil {
			Error(c, http.StatusNotFound, "原报价不存在")
			return
		}
		if original.UserID != userID {
			Error(c, http.StatusNotFound, "原报价不存在")
			return
		}

		newQuote := original
		newQuote.ID = 0
		newQuote.UserID = userID
		newQuote.QuoteNumber = fmt.Sprintf("QT-%s-%03d", time.Now().Format("2006"), time.Now().UnixNano()%1000)
		newQuote.Status = "草稿"
		newQuote.Items = nil

		db.Create(&newQuote)

		for _, item := range original.Items {
			item.ID = 0
			item.QuoteID = newQuote.ID
			db.Create(&item)
		}

		db.Preload("Items").First(&newQuote, newQuote.ID)
		Success(c, newQuote)
	}
}
