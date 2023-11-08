package utils

import (
	"errors"
	"github.com/helloh2o/lucky/log"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandString by len
func RandString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := rand.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}

// RandGroup by []unit32
func RandGroup(p ...uint32) int {
	if p == nil {
		panic("args not found")
	}

	r := make([]uint32, len(p))
	for i := 0; i < len(p); i++ {
		if i == 0 {
			r[0] = p[0]
		} else {
			r[i] = r[i-1] + p[i]
		}
	}

	rl := r[len(r)-1]
	if rl == 0 {
		return 0
	}

	rn := uint32(rand.Int63n(int64(rl)))
	for i := 0; i < len(r); i++ {
		if rn < r[i] {
			return i
		}
	}

	panic("bug")
}

// RandInterval b1 to b2
func RandInterval(b1, b2 int32) int32 {
	if b1 == b2 {
		return b1
	}

	min, max := int64(b1), int64(b2)
	if min > max {
		min, max = max, min
	}
	return int32(rand.Int63n(max-min+1) + min)
}

// RandIntervalN b1, b2, n
func RandIntervalN(b1, b2 int32, n uint32) []int32 {
	if b1 == b2 {
		return []int32{b1}
	}

	min, max := int64(b1), int64(b2)
	if min > max {
		min, max = max, min
	}
	l := max - min + 1
	if int64(n) > l {
		n = uint32(l)
	}

	r := make([]int32, n)
	m := make(map[int32]int32)
	for i := uint32(0); i < n; i++ {
		v := int32(rand.Int63n(l) + min)

		if mv, ok := m[v]; ok {
			r[i] = mv
		} else {
			r[i] = v
		}

		lv := int32(l - 1 + min)
		if v != lv {
			if mv, ok := m[lv]; ok {
				m[v] = mv
			} else {
				m[v] = lv
			}
		}

		l--
	}

	return r
}

// 根据权重随机 RandItemWeight
func RandItemWeight(data map[interface{}]int64) (interface{}, error) {
	items := make(map[interface{}][]int64)
	// 随机
	max := int64(0)
	for item, weight := range data {
		size := weight
		from := max
		to := max + size - 1
		items[item] = []int64{from, to}
		max += size
	}
	if max <= 0 {
		return nil, errors.New("no rand by weight 0")
	}
	// 随机位置
	randWeight := rand.Int63n(max)
	// 概率
	probRecord := float64(randWeight) / float64(max)
	log.Debug("Item rand, index:%d, max:%d, weight:%v", randWeight, max, probRecord)
	// 随机到的物品
	for item, pos := range items {
		if randWeight >= pos[0] && randWeight <= pos[1] {
			return item, nil
		}
	}
	return nil, errors.New("no rand items")
}
