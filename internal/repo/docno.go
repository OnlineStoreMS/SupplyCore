package repo

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"gorm.io/gorm"
)

// nextDailyDocNo 按自然日全局取号（不按租户）。同库多租户不再发出相同 PO/SH/IN/PR 号。
func nextDailyDocNo(db *gorm.DB, model any, col, prefix string) (string, error) {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" || col == "" {
		return "", fmt.Errorf("doc no prefix required")
	}
	var last string
	err := db.Model(model).
		Where(col+" LIKE ?", prefix+"%").
		Order(col + " DESC").
		Limit(1).
		Pluck(col, &last).Error
	if err != nil {
		return "", err
	}
	seq := 1
	if last != "" && len(last) > len(prefix) {
		var n int
		if _, scanErr := fmt.Sscanf(last[len(prefix):], "%d", &n); scanErr == nil && n >= 0 {
			seq = n + 1
		}
	}
	return fmt.Sprintf("%s%04d", prefix, seq), nil
}

func dailyPrefix(code string) string {
	return strings.TrimSpace(code) + time.Now().Format("20060102")
}

// LooksLikeFullPONo 完整代发/采购单号（PO + 至少 12 位数字），用于精确查找而非模糊搜索。
func LooksLikeFullPONo(s string) bool {
	s = strings.ToUpper(strings.TrimSpace(s))
	if !strings.HasPrefix(s, "PO") {
		return false
	}
	rest := s[2:]
	if len(rest) < 12 {
		return false
	}
	for _, r := range rest {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
