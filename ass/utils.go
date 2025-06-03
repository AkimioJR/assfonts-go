package ass

import (
	"strings"
)

// 判断字符串是否有前缀（不区分大小写）
func startWith(raw string, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(raw), strings.ToLower(prefix))
}
