package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"quotepro-backend/config"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
}

func extractJSONPayload(raw string) string {
	s := strings.TrimSpace(raw)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSpace(s)
		if strings.HasPrefix(strings.ToLower(s), "json") {
			s = strings.TrimSpace(s[4:])
		}
		if i := strings.LastIndex(s, "```"); i >= 0 {
			s = strings.TrimSpace(s[:i])
		}
	}
	return strings.TrimSpace(s)
}

// ExtractJSONPayload 去掉 ```json ... ``` 之类包裹，便于解析 JSON。
func ExtractJSONPayload(raw string) string {
	return extractJSONPayload(raw)
}

func chatCompletionsURL(cfg *config.Config) string {
	base := strings.TrimSuffix(strings.TrimSpace(cfg.OpenAIBase), "/")
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	return base + "/chat/completions"
}

func callOpenAI(cfg *config.Config, messages []ChatMessage) (string, error) {
	if cfg.OpenAIAPIKey == "" {
		return "", fmt.Errorf("OpenAI API Key 未配置")
	}

	reqBody := ChatRequest{
		Model:    cfg.OpenAIModel,
		Messages: messages,
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", chatCompletionsURL(cfg), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.OpenAIAPIKey)

	// AI 生成可能耗时较长，这里给一个合理上限，避免连接无限期挂起
	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("请求 OpenAI 失败: %v", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("OpenAI 返回错误 %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("OpenAI 未返回结果")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// ChatWithAI 直接调用 OpenAI-compatible 的 chat/completions，
// 返回 assistant 的文本内容（用于“对话生成询盘”等非 JSON 场景）。
func ChatWithAI(cfg *config.Config, messages []ChatMessage) (string, error) {
	return callOpenAI(cfg, messages)
}

func ParseRequirementWithAI(cfg *config.Config, requirement string) (map[string]interface{}, error) {
	prompt := `你是一个专业的外贸报价助手。请分析以下客户需求，提取结构化参数。

严格规则：
1. 只输出一个 JSON 对象，不要 markdown 代码块、不要解释、不要前后缀文字。
2. 无法从询盘中确定的字段必须留空字符串或 0，禁止编造客户名、价格、MOQ、交期、付款方式等。
3. 支持中英混合询盘。
4. 字段名必须使用下面列出的 camelCase 键名。

JSON 结构：
{
  "customerName": "",
  "country": "",
  "currency": "USD",
  "deliveryAddress": "",
  "productName": "",
  "model": "",
  "material": "",
  "size": "",
  "color": "",
  "quantity": 0,
  "packaging": "",
  "moq": 0,
  "paymentTerms": "",
  "leadTime": "",
  "validityPeriod": "",
  "includeShipping": null,
  "unconfirmed": ["未能从询盘识别的参数中文名列表"]
}

若 AI 曾使用 snake_case 或 missing_fields，也会在服务端归一化；你仍应优先使用上述 camelCase 与 unconfirmed。`

	messages := []ChatMessage{
		{Role: "system", Content: prompt},
		{Role: "user", Content: requirement},
	}

	result, err := callOpenAI(cfg, messages)
	if err != nil {
		return nil, err
	}

	payload := extractJSONPayload(result)
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
		return nil, fmt.Errorf("AI 返回格式错误: %v", err)
	}

	return parsed, nil
}

func GenerateQuoteWithAI(cfg *config.Config, params map[string]interface{}, hints map[string]interface{}) (map[string]interface{}, error) {
	paramsJSON, _ := json.Marshal(params)

	basePrompt := `你是一个专业的外贸 B2B 销售。根据结构化参数生成报价草稿与客户回复。

严格规则：
1. 只输出一个 JSON 对象，不要 markdown、不要解释。
2. 不得承诺询盘与参数中未出现的具体规格、价格、交期；单价/总价可为基于行业常识的参考草稿，须在话术中使用委婉表述（如 "subject to final confirmation" / 「仅供参考，以合同为准」），不得写成绝对承诺。
3. 至少包含 1 条报价行 items；replyVersions 必须 3 条，顺序为：简短成交版、专业邮件版、追单版。
4. 确认清单 confirmationList 为数组；attachments 可为占位文件名列表，selected 布尔值。

JSON 结构：
{
  "items": [
    {"productName": "", "model": "", "specs": "", "quantity": 0, "unitPrice": 0, "totalPrice": 0, "remark": ""}
  ],
  "replyVersions": [
    {"title": "简短成交版 (WhatsApp/微信)", "content": "", "language": "en"},
    {"title": "专业邮件版", "content": "", "language": "en"},
    {"title": "追单版", "content": "", "language": "en"}
  ],
  "confirmationList": [
    {"question": "", "questionEn": "", "checked": false}
  ],
  "attachments": [
    {"name": "Product Spec Sheet.pdf", "url": "", "selected": true}
  ],
  "totalAmount": 0,
  "currency": "USD"
}`

	var extra []string
	if hints != nil {
		if t, ok := hints["_replyTone"].(string); ok && t != "" {
			extra = append(extra, "回复语气偏好："+t+"。\n")
		}
		if lang, ok := hints["_replyLanguage"].(string); ok && lang != "" {
			if lang == "zh" {
				extra = append(extra, "三条 replyVersions 的 content 使用中文。\n")
			} else {
				extra = append(extra, "三条 replyVersions 的 content 使用英文。\n")
			}
		}
		if idx, ok := hints["_replyIndex"]; ok {
			extra = append(extra, fmt.Sprintf("仅重点重写 replyVersions 中索引 %v 对应的那一条（0 起算），其余两条可与输入上下文一致或略作润色，但须输出完整 3 条数组。\n", idx))
		}
	}
	if len(extra) > 0 {
		basePrompt += "\n附加说明：\n" + strings.Join(extra, "")
	}

	messages := []ChatMessage{
		{Role: "system", Content: basePrompt},
		{Role: "user", Content: string(paramsJSON)},
	}

	result, err := callOpenAI(cfg, messages)
	if err != nil {
		return nil, err
	}

	payload := extractJSONPayload(result)
	var generated map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &generated); err != nil {
		return nil, fmt.Errorf("AI 返回格式错误: %v", err)
	}

	return generated, nil
}

// GenerateInquiryExamplesWithAI 按分组主题生成多条「客户询盘」示例，返回 JSON 解析后的 examples 数组。
func GenerateInquiryExamplesWithAI(cfg *config.Config, groupName, groupPrompt, language string, count int) ([]map[string]interface{}, error) {
	if count < 1 {
		count = 3
	}
	if count > 10 {
		count = 10
	}
	lang := strings.ToLower(strings.TrimSpace(language))
	if lang != "zh" {
		lang = "en"
	}

	system := `你是外贸 B2B 场景下的文案助手。根据「分组名称」和「分组说明」生成多条互不重复的客户询盘示例文本，用于粘贴到报价系统的需求输入框。

严格规则：
1) 只输出一个 JSON 对象，不要 markdown、不要解释、不要代码块标记。
2) JSON 结构固定为：{"examples":[{"title":"","lang":"zh"|"en","content":""}]}
3) title 为简短标题；content 为完整询盘正文（可多段，用换行）。
4) 不要编造具体价格、认证编号；可写请报价、MOQ、交期待确认等。
5) 各条 content 在场景上要有差异（不同产品角度或不同信息完整度）。`

	langHint := "英文（每条 examples 的 lang 填 en）"
	if lang == "zh" {
		langHint = "中文（每条 examples 的 lang 填 zh）"
	}
	user := fmt.Sprintf(
		"分组名称：%s\n分组说明（创作依据）：%s\n目标语言：%s\n请生成 %d 条互不重复的示例。",
		strings.TrimSpace(groupName),
		strings.TrimSpace(groupPrompt),
		langHint,
		count,
	)
	if lang == "zh" {
		user += "\n所有 content 使用中文（可夹杂常见英文产品词）。"
	} else {
		user += "\nAll content in English."
	}

	messages := []ChatMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: user},
	}

	result, err := callOpenAI(cfg, messages)
	if err != nil {
		return nil, err
	}

	payload := extractJSONPayload(result)
	var wrapper struct {
		Examples []map[string]interface{} `json:"examples"`
	}
	if err := json.Unmarshal([]byte(payload), &wrapper); err != nil {
		return nil, fmt.Errorf("AI 返回格式错误: %v", err)
	}
	if len(wrapper.Examples) == 0 {
		return nil, fmt.Errorf("AI 未返回 examples")
	}
	return wrapper.Examples, nil
}
