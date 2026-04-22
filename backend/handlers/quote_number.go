package handlers

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// allocateNextQuoteNumber 生成 QYYYYMMDD-0001 形式单号（按当日最大序号 +1）。
func allocateNextQuoteNumber(db *gorm.DB) string {
	prefix := "Q" + time.Now().Format("20060102") + "-"
	var maxSeq int64
	_ = db.Raw(
		`SELECT COALESCE(MAX(CAST(SUBSTR(quote_number, 11) AS INTEGER)), 0) FROM quotes WHERE quote_number LIKE ? AND deleted_at IS NULL`,
		prefix+"%",
	).Scan(&maxSeq)
	return fmt.Sprintf("%s%04d", prefix, maxSeq+1)
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "unique") || strings.Contains(s, "constraint")
}
