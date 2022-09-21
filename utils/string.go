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

func IsEmail(email string) bool {
	if email != "" {
		if isOk, _ := regexp.MatchString("^[_a-z0-9-]+(\\.[_a-z0-9-]+)*@[a-z0-9-]+(\\.[a-z0-9-]+)*(\\.[a-z]{2,4})$", email); isOk {
			return true
		}
	}
	return false
}

func IsPhone(phoneStr string) bool {
	if phoneStr != "" {
		if isOk, _ := regexp.MatchString(`^\([\d]{3}\) [\d]{3}-[\d]{4}$`, phoneStr); isOk {
			return isOk
		}
	}

	return false
}
