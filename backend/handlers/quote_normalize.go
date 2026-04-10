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

func normalizeGenerateResponse(raw map[string]interface{}, fallbackCustomer, fallbackCurrency, replyLang string) map[string]interface{} {
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
	out["replyVersions"] = normalizeReplyVersions(rv, replyLang)

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

func normalizeReplyVersions(rv interface{}, replyLang string) []map[string]interface{} {
	if m, ok := rv.(map[string]interface{}); ok {
		lang := "en"
		if replyLang == "zh" {
			lang = "zh"
		}
		return []map[string]interface{}{
			{"title": "简短成交版 (WhatsApp/微信)", "content": strVal(m, "short"), "language": lang},
			{"title": "专业邮件版", "content": strVal(m, "professional"), "language": lang},
			{"title": "追单版", "content": strVal(m, "followup"), "language": lang},
		}
	}
	arr := toIfaceSlice(rv)
	var out []map[string]interface{}
	for _, x := range arr {
		row, ok := x.(map[string]interface{})
		if !ok {
			continue
		}
		lang := strVal(row, "language")
		if lang == "" {
			lang = "en"
		}
		out = append(out, map[string]interface{}{
			"title":   strVal(row, "title"),
			"content": strVal(row, "content"),
			"language": lang,
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
