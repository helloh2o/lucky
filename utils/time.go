package utils

import "time"

// FormatTime2String 时间to字符串
func FormatTime2String(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// GetTomorrowUnix 获取明日凌晨时间戳s
func GetTomorrowUnix() int64 {
	return time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local).Unix() + 86400
}

// GetTodayUnix 获取今日凌晨时间戳s
func GetTodayUnix() int64 {
	return time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Local).Unix()
}
