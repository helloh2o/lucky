package utils

import "time"

// FormatTime2String 时间to字符串
func FormatTime2String(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
