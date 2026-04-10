package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"quotepro-backend/config"
	"quotepro-backend/models"
	"quotepro-backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GenerateTemplateByCategory 用 AI 根据分类名称生成一份模板并保存到数据库。
// 多语言：language=zh/en。
func GenerateTemplateByCategory(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		var body struct {
			Category      string `json:"category"`
			CategoryLabel string `json:"categoryLabel"`
			Language      string `json:"language"`
			NameHint      string `json:"nameHint"`
			ExtraHint     string `json:"extraHint"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			Error(c, http.StatusBadRequest, "参数错误")
			return
		}

		category := strings.TrimSpace(body.Category)
		if category == "" {
			Error(c, http.StatusBadRequest, "请提供分类")
			return
		}
		catLabel := strings.TrimSpace(body.CategoryLabel)
		if catLabel == "" {
			catLabel = category
		}
		lang := strings.ToLower(strings.TrimSpace(body.Language))
		if lang != "zh" && lang != "en" {
			lang = "zh"
		}
		nameHint := strings.TrimSpace(body.NameHint)
		extraHint := strings.TrimSpace(body.ExtraHint)

		varVar := `可用变量（用 {{var}} 插入）：{{customer_name}}, {{product_name}}, {{quantity}}, {{price}}, {{lead_time}}, {{payment_terms}}`

		system := `你是外贸 B2B 模板专家。根据给定「模板分类」生成一份可直接使用的模板内容。

严格规则：
1) 只输出一个 JSON 对象，不要 markdown，不要解释，不要代码块。
2) JSON 结构固定为：{"name":"模板名称","content":"模板正文"}。
3) content 中必须合理使用变量占位符：{{customer_name}}、{{product_name}}、{{quantity}}、{{price}}、{{lead_time}}、{{payment_terms}}（按需要使用，不要全部硬塞）。
4) 语气真实、可直接发给客户；不要编造具体价格数字或不存在的认证编号。`
		if lang == "zh" {
			system += "\n5) 输出中文。"
		} else {
			system += "\n5) Output in English."
		}

		user := fmt.Sprintf("模板分类：%s\n分类说明：%s\n%s\n", category, catLabel, varVar)
		if nameHint != "" {
			user += "模板命名偏好：" + nameHint + "\n"
		}
		if extraHint != "" {
			user += "补充说明（必须尽量满足）：\n" + extraHint + "\n"
		}
		// 按类别加一点结构约束，避免“空内容”
		switch category {
		case "quotation":
			user += "要求：content 以报价单/报价说明为主，包含：产品概述、价格说明（可写 subject to final confirmation）、交期、付款条款、有效期、包装/运输说明、结尾礼貌收尾。\n"
		case "email":
			user += "要求：content 是专业邮件回复模板，包含称呼、感谢、报价要点、待确认项（可选）、结尾签名。\n"
		case "chat":
			user += "要求：content 是简短即时通讯话术（WhatsApp/微信风格），语气更口语，但仍专业；不要太长。\n"
		case "confirmation":
			user += "要求：content 是参数确认清单模板（面向客户），用简短条目询问缺失信息；可用中英双语或按 language 输出。\n"
		default:
			user += "要求：content 与该分类语境匹配，结构清晰，可直接复制发送。\n"
		}

		messages := []services.ChatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		}

		raw, err := services.ChatWithAI(cfg, messages)
		if err != nil {
			Error(c, http.StatusInternalServerError, "AI 生成失败: "+err.Error())
			return
		}

		payload := services.ExtractJSONPayload(raw)
		var out struct {
			Name    string `json:"name"`
			Content string `json:"content"`
		}
		if err := json.Unmarshal([]byte(payload), &out); err != nil {
			Error(c, http.StatusInternalServerError, "AI 返回格式错误: "+err.Error())
			return
		}
		out.Name = strings.TrimSpace(out.Name)
		out.Content = strings.TrimSpace(out.Content)
		if out.Name == "" {
			out.Name = fmt.Sprintf("%s-%s", catLabel, map[bool]string{true: "中文", false: "EN"}[lang == "zh"])
		}
		if out.Content == "" {
			Error(c, http.StatusInternalServerError, "AI 返回内容为空")
			return
		}

		tmpl := models.Template{
			UserID:   userID,
			Name:     out.Name,
			Category: category,
			Language: lang,
			Content:  out.Content,
		}
		if err := db.Create(&tmpl).Error; err != nil {
			Error(c, http.StatusInternalServerError, "创建模板失败")
			return
		}
		Success(c, tmpl)
	}
}

