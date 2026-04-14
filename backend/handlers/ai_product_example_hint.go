package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"quotepro-backend/config"
	"quotepro-backend/services"

	"github.com/gin-gonic/gin"
)

// ComposeProductExampleHint 生成“产品示例 AI 生成”的补充说明文本（给用户二次编辑）。
// 约定：不要求 JSON，只返回纯文本 content。
func ComposeProductExampleHint(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			Count       int    `json:"count"`
			CurrentHint string `json:"currentHint"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			Error(c, http.StatusBadRequest, "参数错误")
			return
		}
		count := body.Count
		if count < 1 {
			count = 3
		}
		if count > 5 {
			count = 5
		}
		cur := strings.TrimSpace(body.CurrentHint)

		system := `你是外贸 B2B SaaS 的产品资料库助手。你的任务是帮用户生成一段“补充说明”，用于指导 AI 生成可导入的产品示例数据。

输出要求：
1) 只输出纯文本（中文为主，可夹杂常见英文缩写），不要 JSON，不要 markdown，不要编号列表的前后缀说明。
2) 内容应可直接粘贴到“补充说明”输入框，长度控制在 1-4 行。
3) 说明要尽量具体：品类范围/字段偏好/允许为空/价格与MOQ策略/交期付款等。
4) 不要编造公司名/客户名等无关信息。`

		user := "请生成一段用于“AI 生成产品示例”的补充说明。\n"
		user += "条数上限：5；本次条数：" + fmt.Sprint(count) + "\n"
		if cur != "" {
			user += "用户当前草稿（可基于它优化、补全，但不要重复啰嗦）：\n" + cur + "\n"
		}
		user += "尽量覆盖：品类、字段（name/sku/category/material/size/color/process/packaging/price/moq/leadTime/paymentTerms/description），并提示哪些可以为空。"

		messages := []services.ChatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		}
		out, err := services.ChatWithAI(cfg, messages)
		if err != nil {
			Error(c, http.StatusInternalServerError, "AI 生成失败: "+err.Error())
			return
		}
		Success(c, gin.H{"content": strings.TrimSpace(out)})
	}
}

