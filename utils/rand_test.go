package utils

import (
	"github.com/helloh2o/lucky/log"
	"testing"
)

func TestRandItemWeight(t *testing.T) {
	data := map[interface{}]int64{
		"A": 4, // weight
		"B": 3,
		"C": 1,
		"D": 2,
	}
	at := 0
	bt := 0
	ct := 0
	dt := 0
	for i := 0; i < 1000; i++ {
		if item, err := RandItemWeight(data); err == nil {
			log.Debug("item:%v", item)
			switch item.(string) {
			case "A":
				at++
			case "B":
				bt++
			case "C":
				ct++
			case "D":
				dt++
			}
		}
	}
	log.Debug("A prob :%.2f", float64(at)/100.0)
	log.Debug("B prob :%.2f", float64(bt)/100.0)
	log.Debug("C prob :%.2f", float64(ct)/100.0)
	log.Debug("D prob :%.2f", float64(dt)/100.0)
}
