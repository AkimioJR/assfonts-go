package srt

import "strings"

// 辅助函数：转换时间格式为ASS需要的格式 (HH:MM:SS.ff)
func convertTime(s string) string {
	s = strings.Replace(s, ",", ".", 1) // 将逗号替换为句点
	parts := strings.Split(s, ".")
	if len(parts) < 2 {
		return s + ".00"
	}

	// 处理毫秒部分，确保是2位
	millis := parts[1]
	if len(millis) > 2 {
		millis = millis[:2]
	} else if len(millis) < 2 {
		millis = millis + strings.Repeat("0", 2-len(millis))
	}

	return parts[0] + "." + millis
}
