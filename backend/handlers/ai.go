package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"quotepro-backend/config"
	"quotepro-backend/services"

	"github.com/gin-gonic/gin"
)

// ComposeInquiry 用“对话收集到的信息”生成一段可粘贴到询盘输入框的文本。
// 约定：不要求 JSON，只返回纯文本 content。
func ComposeInquiry(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			Language string            `json:"language"`
			Answers  map[string]string `json:"answers"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			Error(c, http.StatusBadRequest, "参数错误")
			return
		}

		lang := strings.ToLower(strings.TrimSpace(body.Language))
		if lang != "zh" && lang != "en" {
			lang = "en"
		}
		if len(body.Answers) == 0 {
			Error(c, http.StatusBadRequest, "请先填写对话信息")
			return
		}

		// 把 answers 拼成可读上下文，让模型输出一段可直接粘贴的“客户询盘文本”
		var lines []string
		for k, v := range body.Answers {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			lines = append(lines, k+": "+v)
		}
		ctx := strings.Join(lines, "\n")

		system := `你是一个专业的外贸销售助手。你的任务是根据用户提供的信息，生成一段“客户发来的询盘文本”，用于粘贴到系统的客户需求输入框。

严格规则：
1) 只输出询盘正文纯文本，不要 markdown，不要前后缀说明，不要编号清单。
2) 不确定的信息不要编造，可以用委婉表述或留空不写。
3) 内容要像真实客户写来的询盘：包含产品、数量、交付地、包装/材质/尺寸等（若有）。
4) 语气自然，B2B 采购沟通风格。`
		if lang == "zh" {
			system += "\n5) 输出使用中文。"
		} else {
			system += "\n5) Output in English."
		}

		messages := []services.ChatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: ctx},
		}

		out, err := services.ChatWithAI(cfg, messages)
		if err != nil {
			Error(c, http.StatusInternalServerError, "AI 生成失败: "+err.Error())
			return
		}
		Success(c, gin.H{"content": strings.TrimSpace(out)})
	}
}

// GenerateInquiryExamples 按分组名称 + 分组提示词批量生成询盘示例（JSON）。
func GenerateInquiryExamples(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			GroupName   string `json:"groupName"`
			GroupPrompt string `json:"groupPrompt"`
			Language    string `json:"language"`
			Count       int    `json:"count"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			Error(c, http.StatusBadRequest, "参数错误")
			return
		}
		name := strings.TrimSpace(body.GroupName)
		if name == "" {
			Error(c, http.StatusBadRequest, "请填写分组名称")
			return
		}
		prompt := strings.TrimSpace(body.GroupPrompt)
		if prompt == "" {
			Error(c, http.StatusBadRequest, "请填写分组提示词（用于 AI 理解本组主题）")
			return
		}
		lang := strings.ToLower(strings.TrimSpace(body.Language))
		if lang != "zh" && lang != "en" {
			lang = "en"
		}
		count := body.Count
		if count < 1 {
			count = 3
		}
		if count > 10 {
			count = 10
		}

		rows, err := services.GenerateInquiryExamplesWithAI(cfg, name, prompt, lang, count)
		if err != nil {
			Error(c, http.StatusInternalServerError, "AI 生成失败: "+err.Error())
			return
		}

		var out []gin.H
		for _, row := range rows {
			title := strings.TrimSpace(fmt.Sprint(row["title"]))
			l := strings.ToLower(strings.TrimSpace(fmt.Sprint(row["lang"])))
			if l != "zh" && l != "en" {
				l = lang
			}
			content := strings.TrimSpace(fmt.Sprint(row["content"]))
			if content == "" {
				continue
			}
			out = append(out, gin.H{
				"title":   title,
				"lang":    l,
				"content": content,
			})
		}
		if len(out) == 0 {
			Error(c, http.StatusInternalServerError, "AI 未生成有效示例")
			return
		}
		Success(c, gin.H{"examples": out})
	}
}

