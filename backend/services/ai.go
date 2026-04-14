package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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

【纺织品/面料类 — 易错点，务必遵守】
- material：材质与布面规格。含「克重」「gsm」「g/m²」等面布克重信息时，必须写入本字段（例如「约280gsm」或「棉/涤，280gsm」），不要写进 size。
- size：物理尺寸/外形尺寸。含「门幅」「幅宽」「宽度」以及「150cm」「1.5m」等长度/宽度表述时写入本字段（例如「门幅150cm」）。不要把「280gsm」「克重」仅填在 size 里。
- 若一句同时出现克重与门幅：克重相关归 material；门幅/幅宽/cm 归 size。可拆成两段文字分别填入两字段。

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
	if err := ValidateParsedRequirementShape(parsed); err != nil {
		return nil, err
	}

	return parsed, nil
}

func GenerateQuoteWithAI(cfg *config.Config, params map[string]interface{}, hints map[string]interface{}) (map[string]interface{}, error) {
	paramsJSON, _ := json.Marshal(params)

	basePrompt := `你是一个专业的外贸 B2B 销售。根据结构化参数生成报价草稿与客户回复。

严格规则：
1. 只输出一个 JSON 对象，不要 markdown、不要解释。
2. 不得承诺或编造用户 JSON 中**未出现**的规格、价格、交期；若用户 JSON 里**已有**某规格/数量/价格草稿，则三条 replyVersions（含简短版）都必须**写出该具体值**，句末可再加委婉限制（如 subject to final confirmation），**禁止**用一句笼统的「均以最终确认为准」完全顶替用户已填信息而不出现任何数字或要点。单价/总价可为行业常识参考草稿，同样须带委婉表述，不得写成绝对承诺。
3. 至少包含 1 条报价行 items；replyVersions 必须 3 条，顺序为：简短成交版、专业邮件版、追单版。
4. 确认清单 confirmationList 为数组；attachments 可为占位文件名列表，selected 布尔值。
5. 【必须落地用户已填参数】用户 JSON 里**非空**的字段必须在 replyVersions 正文中自然出现：称呼用 customerName；交付地/目的港用 deliveryAddress；国家用 country；产品用 productName、model、material、size、color、quantity、packaging、moq、paymentTerms、leadTime 等对应键的实际值。不要把已填信息再写成「待确认」清单里的重复项；未提供的字段不要编造。
6. 【禁止卖方占位符】三条 replyVersions 的 content 中禁止使用 [Your Full Name]、[Your Company Name]、[Contact Info]、[Your Position] 等方括号占位。若无卖方具体信息，结尾用「Best regards,」加一行「Sales Team」或仅「Best regards,」即可。
7. confirmationList 优先列出用户 JSON 中仍为空、或出现在 unconfirmed 里、或解析未覆盖的要点；避免把 customerName、deliveryAddress、productName 等已填字段再当成「待客户补充」重复提问。
8. 【简短成交版不得整段套话】replyVersions[0]（简短成交版）可短句、口语、可用 emoji，但必须点到用户已填的**至少若干项具体信息**（如 material、size 内 gsm/门幅、color、quantity、deliveryAddress/目的港、moq 等），**禁止**全段只有「Price, MOQ, T-count, payment etc. all subject to confirmation」类空话而完全不出现已填字段的具体值。

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
	if err := ValidateGenerateQuoteShape(generated); err != nil {
		return nil, err
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

// GenerateProductExamplesWithAI 生成可直接导入产品库的示例产品列表。
// count 最大 5；extraHint 为可选补充说明（例如品类、材质、规格字段偏好等）。
func GenerateProductExamplesWithAI(cfg *config.Config, count int, extraHint string) ([]map[string]interface{}, error) {
	if count < 1 {
		count = 3
	}
	if count > 5 {
		count = 5
	}
	extraHint = strings.TrimSpace(extraHint)

	system := `你是外贸 B2B 产品资料库助手。请生成若干条“产品资料库”示例数据，供卖家导入自己的产品库。

严格规则：
1) 只输出一个 JSON 对象，不要 markdown、不要解释、不要代码块。
2) JSON 结构固定为：{"items":[{...},{...}]}
3) items 数组长度必须等于用户指定条数。
4) 字段必须使用 camelCase：name, sku, category, description, material, size, color, process, packaging, price, moq, leadTime, paymentTerms
5) 内容要像真实可售产品资料：name 不要太泛；sku 可为空或短编码；category 用简短类目词；price/moq 用合理参考值（可为 0 表示未知）；leadTime/paymentTerms 可为常见表达或空。
6) 不要输出 attachments/userId/id/createdAt/updatedAt 等系统字段。`

	user := "请生成产品示例。\n"
	user += fmt.Sprintf("条数：%d\n", count)
	if extraHint != "" {
		user += "补充说明：" + extraHint + "\n"
	}
	user += "注意：字段允许为空字符串，但必须给齐上述字段键名。"

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
		Items []map[string]interface{} `json:"items"`
	}
	if err := json.Unmarshal([]byte(payload), &wrapper); err != nil {
		return nil, fmt.Errorf("AI 返回格式错误: %v", err)
	}
	if len(wrapper.Items) != count {
		if len(wrapper.Items) == 0 {
			return nil, fmt.Errorf("AI 未返回 items")
		}
		// 容错：如果返回条数不一致，截断到前 count 条
		if len(wrapper.Items) > count {
			wrapper.Items = wrapper.Items[:count]
		}
	}

	// 归一化字段（确保前端/导入端拿到的结构稳定）
	out := make([]map[string]interface{}, 0, len(wrapper.Items))
	for _, it := range wrapper.Items {
		row := map[string]interface{}{
			"name":         strings.TrimSpace(fmt.Sprint(it["name"])),
			"sku":          strings.TrimSpace(fmt.Sprint(it["sku"])),
			"category":     strings.TrimSpace(fmt.Sprint(it["category"])),
			"description":  strings.TrimSpace(fmt.Sprint(it["description"])),
			"material":     strings.TrimSpace(fmt.Sprint(it["material"])),
			"size":         strings.TrimSpace(fmt.Sprint(it["size"])),
			"color":        strings.TrimSpace(fmt.Sprint(it["color"])),
			"process":      strings.TrimSpace(fmt.Sprint(it["process"])),
			"packaging":    strings.TrimSpace(fmt.Sprint(it["packaging"])),
			"price":        floatValLoose(it["price"]),
			"moq":          intValLoose(it["moq"]),
			"leadTime":     strings.TrimSpace(fmt.Sprint(it["leadTime"])),
			"paymentTerms": strings.TrimSpace(fmt.Sprint(it["paymentTerms"])),
		}
		out = append(out, row)
	}
	return out, nil
}

func floatValLoose(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case json.Number:
		f, _ := t.Float64()
		return f
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		if err == nil {
			return f
		}
	}
	return 0
}

func intValLoose(v interface{}) int {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case float32:
		return int(t)
	case json.Number:
		i, _ := t.Int64()
		return int(i)
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(t))
		if err == nil {
			return n
		}
	}
	return 0
}
