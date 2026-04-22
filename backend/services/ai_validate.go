package services

import "fmt"

func toIfaceSlice(v interface{}) []interface{} {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case []interface{}:
		return t
	case []map[string]interface{}:
		out := make([]interface{}, len(t))
		for i := range t {
			out[i] = t[i]
		}
		return out
	default:
		return nil
	}
}

func isStringishSlice(v interface{}) bool {
	switch t := v.(type) {
	case []string:
		return true
	case []interface{}:
		for _, x := range t {
			if x == nil {
				continue
			}
			if _, ok := x.(string); !ok {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// ValidateParsedRequirementShape 校验解析接口返回的最小结构，避免脏数据进入后续流程。
func ValidateParsedRequirementShape(m map[string]interface{}) error {
	if m == nil || len(m) == 0 {
		return fmt.Errorf("解析结果为空")
	}
	if u, ok := m["unconfirmed"]; ok && u != nil {
		if !isStringishSlice(u) {
			return fmt.Errorf("unconfirmed 必须为字符串数组")
		}
	}
	return nil
}

// ValidateGenerateQuoteShape 校验报价生成 JSON：至少 1 行 items、3 条 replyVersions。
func ValidateGenerateQuoteShape(m map[string]interface{}) error {
	if m == nil {
		return fmt.Errorf("生成结果为空")
	}
	items := toIfaceSlice(m["items"])
	if len(items) < 1 {
		return fmt.Errorf("AI 输出缺少报价行 items")
	}
	rv := toIfaceSlice(m["replyVersions"])
	if len(rv) != 3 {
		return fmt.Errorf("replyVersions 必须为 3 条，当前 %d 条", len(rv))
	}
	return nil
}
