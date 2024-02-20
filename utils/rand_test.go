package utils

import (
	"math/rand"
	"testing"
)

func TestRandItemWeight(t *testing.T) {
	rd := rand.New(rand.NewSource(100))
	data := map[interface{}]int64{
		1: 0,
		2: 0,
		3: 0,
	}
	n1, n2, n3 := 0, 0, 0
	for i := 0; i < 10; i++ {
		item, err := RandItemWeight(rd, data)
		if err != nil {
			t.Error(err)
		}
		if item == 1 {
			n1++
		}
		if item == 2 {
			n2++
		}
		if item == 3 {
			n3++
		}
		t.Log(item)
	}
	t.Log(n1, n2, n3, n1+n2+n3)
}
