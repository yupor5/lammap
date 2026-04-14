package handlers

import (
	"fmt"
	"strconv"
	"strings"
)

var parseSnakeToCamel = map[string]string{
	"customer_name":    "customerName",
	"delivery_place":   "deliveryAddress",
	"product_name":     "productName",
	"payment_term":     "paymentTerms",
	"validity":         "validityPeriod",
	"include_shipping": "includeShipping",
	"missing_fields":   "_missing_fields",
}

func normalizeParsedParams(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return map[string]interface{}{}
	}
	out := make(map[string]interface{})
	for k, v := range m {
		if ck, ok := parseSnakeToCamel[k]; ok {
			if ck == "_missing_fields" {
				mergeStringSlice(out, "unconfirmed", v)
				continue
			}
			out[ck] = v
			continue
		}
		out[k] = v
	}
	mergeStringSlice(out, "unconfirmed", m["missing_fields"])
	if u, ok := out["unconfirmed"].([]interface{}); ok {
		var strs []string
		for _, x := range u {
			if s, ok := x.(string); ok && s != "" {
				strs = append(strs, s)
			}
		}
		out["unconfirmed"] = strs
	}
	return out
}

func mergeStringSlice(out map[string]interface{}, key string, v interface{}) {
	if v == nil {
		return
	}
	arr, ok := v.([]interface{})
	if !ok {
		return
	}
	existing, _ := out[key].([]interface{})
	out[key] = append(existing, arr...)
}

// inferReplyLangFromContent 根据正文是否含足够汉字判断展示语言，避免 AI 在 JSON 里写 language:"en" 却输出中文导致标签与正文不一致。
func inferReplyLangFromContent(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "en"
	}
	var cjk int
	for _, r := range s {
		if r >= 0x4E00 && r <= 0x9FFF {
			cjk++
		}
	}
	if cjk >= 2 {
		return "zh"
	}
	return "en"
}

func splitQuoteMeta(params map[string]interface{}) (clean map[string]interface{}, hints map[string]interface{}) {
	clean = make(map[string]interface{})
	hints = make(map[string]interface{})
	for k, v := range params {
		if strings.HasPrefix(k, "_") {
			hints[k] = v
			continue
		}
		clean[k] = v
	}
	return clean, hints
}

func strVal(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if m == nil {
			break
		}
		if v, ok := m[k]; ok {
			switch t := v.(type) {
			case string:
				return t
			case fmt.Stringer:
				return t.String()
			default:
				return fmt.Sprint(t)
			}
		}
	}
	return ""
}

func floatVal(v interface{}) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int:
		return float64(t)
	case int64:
		return float64(t)
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		if err == nil {
			return f
		}
	}
	return 0
}

func intVal(v interface{}) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(t))
		if err == nil {
			return n
		}
	}
	return 0
}

func normalizeGenerateResponse(raw map[string]interface{}, fallbackCustomer, fallbackCurrency string) map[string]interface{} {
	out := make(map[string]interface{})
	itemsIn := toIfaceSlice(raw["items"])
	if len(itemsIn) == 0 {
		itemsIn = toIfaceSlice(raw["quote_items"])
	}
	var items []map[string]interface{}
	var sum float64
	for _, it := range itemsIn {
		row, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		qty := intVal(row["quantity"])
		up := floatVal(row["unitPrice"])
		if up == 0 {
			up = floatVal(row["unit_price"])
		}
		tp := floatVal(row["totalPrice"])
		if tp == 0 {
			tp = floatVal(row["amount"])
		}
		if tp == 0 && qty > 0 && up > 0 {
			tp = float64(qty) * up
		}
		sum += tp
		items = append(items, map[string]interface{}{
			"productName": strVal(row, "productName", "product_name"),
			"model":       strVal(row, "model"),
			"specs":       strVal(row, "specs", "spec"),
			"quantity":    qty,
			"unitPrice":   up,
			"totalPrice":  tp,
			"remark":      strVal(row, "remark"),
		})
	}
	out["items"] = items

	rv := raw["replyVersions"]
	if rv == nil {
		rv = raw["reply_versions"]
	}
	out["replyVersions"] = normalizeReplyVersions(rv)

	out["confirmationList"] = normalizeConfirmationList(raw["confirmationList"])
	if out["confirmationList"] == nil {
		out["confirmationList"] = normalizeConfirmationList(raw["confirmation_list"])
	}

	att := raw["attachments"]
	if att == nil {
		att = raw["attachment_list"]
	}
	out["attachments"] = normalizeAttachments(att)

	ta := floatVal(raw["totalAmount"])
	if ta == 0 {
		ta = sum
	}
	out["totalAmount"] = ta

	cur := strVal(raw, "currency")
	if cur == "" {
		cur = fallbackCurrency
	}
	if cur == "" {
		cur = "USD"
	}
	out["currency"] = cur

	cn := strVal(raw, "customerName", "customer_name")
	if cn == "" {
		cn = fallbackCustomer
	}
	out["customerName"] = cn
	out["status"] = "草稿"
	return out
}

func normalizeReplyVersions(rv interface{}) []map[string]interface{} {
	if m, ok := rv.(map[string]interface{}); ok {
		short := strVal(m, "short")
		prof := strVal(m, "professional")
		follow := strVal(m, "followup")
		return []map[string]interface{}{
			{"title": "简短成交版 (WhatsApp/微信)", "content": short, "language": inferReplyLangFromContent(short)},
			{"title": "专业邮件版", "content": prof, "language": inferReplyLangFromContent(prof)},
			{"title": "追单版", "content": follow, "language": inferReplyLangFromContent(follow)},
		}
	}
	arr := toIfaceSlice(rv)
	var out []map[string]interface{}
	for _, x := range arr {
		row, ok := x.(map[string]interface{})
		if !ok {
			continue
		}
		content := strVal(row, "content")
		out = append(out, map[string]interface{}{
			"title":    strVal(row, "title"),
			"content":  content,
			"language": inferReplyLangFromContent(content),
		})
	}
	for len(out) < 3 {
		out = append(out, map[string]interface{}{"title": "", "content": "", "language": "en"})
	}
	return out[:3]
}

func normalizeConfirmationList(v interface{}) []map[string]interface{} {
	arr := toIfaceSlice(v)
	var out []map[string]interface{}
	for _, x := range arr {
		row, ok := x.(map[string]interface{})
		if !ok {
			continue
		}
		checked := false
		if b, ok := row["checked"].(bool); ok {
			checked = b
		}
		out = append(out, map[string]interface{}{
			"question":   strVal(row, "question"),
			"questionEn": strVal(row, "questionEn", "question_en"),
			"checked":    checked,
		})
	}
	return out
}

func normalizeAttachments(v interface{}) []map[string]interface{} {
	arr := toIfaceSlice(v)
	var out []map[string]interface{}
	for _, x := range arr {
		row, ok := x.(map[string]interface{})
		if !ok {
			continue
		}
		sel := true
		if b, ok := row["selected"].(bool); ok {
			sel = b
		}
		out = append(out, map[string]interface{}{
			"name":     strVal(row, "name"),
			"url":      strVal(row, "url", "path"),
			"selected": sel,
		})
	}
	return out
}

func toIfaceSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	if a, ok := v.([]interface{}); ok {
		return a
	}
	return nil
}

// repairTextileSizeMaterial 将面料类询盘中「克重/gsm」与「门幅/幅宽/cm」正确归到 material 与 size，避免尺寸字段被克重占满。
func repairTextileSizeMaterial(out map[string]interface{}) {
	if out == nil {
		return
	}
	raw := strings.TrimSpace(anyString(out["size"]))
	if raw == "" {
		return
	}
	mat := strings.TrimSpace(anyString(out["material"]))

	// 连写无逗号：「…克重…280gsm…门幅150cm」类
	if before, after, ok := splitBeforeFabricWidth(raw); ok && fabricHasWeight(raw) {
		before = strings.TrimSpace(before)
		after = strings.TrimSpace(after)
		if before != "" && isFabricWeightSpec(before) {
			mat = mergeMatLine(mat, before)
			out["material"] = mat
		}
		if after != "" {
			out["size"] = after
		} else {
			out["size"] = ""
		}
		return
	}

	segs := splitFabricSegments(raw)
	if len(segs) > 1 {
		var wts, dims []string
		for _, s := range segs {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			w := isFabricWeightSpec(s)
			d := isFabricDimensionSpec(s)
			switch {
			case w && !d:
				wts = append(wts, s)
			case d:
				dims = append(dims, s)
			default:
				dims = append(dims, s)
			}
		}
		if len(wts) > 0 {
			out["material"] = mergeMatLine(mat, strings.Join(wts, "，"))
		}
		if len(dims) > 0 {
			out["size"] = strings.Join(dims, "，")
		} else {
			out["size"] = ""
		}
		return
	}

	// 单段：仅克重/gsm、无门幅尺寸 → 归入 material，清空 size
	if len(segs) == 1 && isFabricWeightSpec(raw) && !isFabricDimensionSpec(raw) {
		out["material"] = mergeMatLine(mat, raw)
		out["size"] = ""
	}
}

func anyString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return fmt.Sprint(t)
	}
}

func mergeMatLine(existing, frag string) string {
	existing = strings.TrimSpace(existing)
	frag = strings.TrimSpace(frag)
	if frag == "" {
		return existing
	}
	if existing == "" {
		return frag
	}
	if strings.Contains(existing, frag) {
		return existing
	}
	return existing + "，" + frag
}

func splitFabricSegments(s string) []string {
	s = strings.ReplaceAll(s, "；", ",")
	cur := strings.FieldsFunc(s, func(r rune) bool {
		return r == '，' || r == ',' || r == ';'
	})
	var parts []string
	for _, p := range cur {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	if len(parts) == 0 {
		return []string{s}
	}
	return parts
}

// splitBeforeFabricWidth 在「门幅/幅宽」前切开（用于无标点连写，如 克重约280gsm门幅150cm）。
func splitBeforeFabricWidth(raw string) (before, after string, ok bool) {
	for _, kw := range []string{"门幅", "幅宽"} {
		if idx := strings.Index(raw, kw); idx > 0 {
			return raw[:idx], raw[idx:], true
		}
	}
	return "", "", false
}

func fabricHasWeight(s string) bool {
	sl := strings.ToLower(s)
	return strings.Contains(sl, "gsm") || strings.Contains(s, "克重") || strings.Contains(sl, "g/m²") || strings.Contains(sl, "g/m2")
}

func isFabricWeightSpec(s string) bool {
	sl := strings.ToLower(s)
	if strings.Contains(sl, "gsm") {
		return true
	}
	if strings.Contains(s, "克重") {
		return true
	}
	if strings.Contains(sl, "g/") && strings.Contains(sl, "m") && strings.Contains(sl, "²") {
		return true
	}
	return false
}

func isFabricDimensionSpec(s string) bool {
	if strings.Contains(s, "门幅") || strings.Contains(s, "幅宽") {
		return true
	}
	sl := strings.ToLower(s)
	if strings.Contains(sl, "cm") || strings.Contains(sl, "mm") || strings.Contains(sl, "inch") || strings.Contains(s, "英寸") {
		return true
	}
	return false
}
