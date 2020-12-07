package little

import (
	"sort"
	"testing"
)

func (password *password) Len() int {
	return passwordLength
}

func (password *password) Less(i, j int) bool {
	return password[i] < password[j]
}

func (password *password) Swap(i, j int) {
	password[i], password[j] = password[j], password[i]
}

func TestRandPassword(t *testing.T) {
	password := RandPassword()
	t.Log(password)
	bsPassword, err := ParsePassword(password)
	if err != nil {
		t.Error(err)
	}
	sort.Sort(bsPassword)
	for i := 0; i < passwordLength; i++ {
		if bsPassword[i] != byte(i) {
			t.Error("不能出现任何一个重复的byte位，必须由 0-255 组成，并且都需要包含")
		}
	}
}
