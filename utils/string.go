package utils

import (
	"regexp"
	"strings"
)

// ReplaceKeyword 替换关键字
func ReplaceKeyword(cfg, value string) string {
	origin := cfg
	const reg = "(.*)"
	cfg = strings.Replace(cfg, "*", reg, -1)
	r, _ := regexp.Compile(cfg)
	if r.MatchString(value) {
		hide := strings.Replace(origin, "*", "", -1)
		for _, s := range []rune(hide) {
			value = strings.Replace(value, string(s), "*", -1)
		}
	}
	return value
}
