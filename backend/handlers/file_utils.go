package handlers

import "strings"

func safeFileBase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	// 保留中英文、数字、常见符号，其余替换为下划线
	repl := make([]rune, 0, len([]rune(s)))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			repl = append(repl, r)
			continue
		}
		if r == '-' || r == '_' || r == '.' || r == ' ' {
			repl = append(repl, r)
			continue
		}
		// 基本汉字范围（足够用于文件名展示）
		if r >= 0x4E00 && r <= 0x9FFF {
			repl = append(repl, r)
			continue
		}
		repl = append(repl, '_')
	}
	out := strings.TrimSpace(string(repl))
	out = strings.ReplaceAll(out, " ", "_")
	for strings.Contains(out, "__") {
		out = strings.ReplaceAll(out, "__", "_")
	}
	if len(out) > 80 {
		out = out[:80]
	}
	out = strings.Trim(out, "._-")
	return out
}

