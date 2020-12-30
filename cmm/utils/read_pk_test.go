package utils

import (
	"fmt"
	"github.com/helloh2o/lucky/log"
	"strconv"
	"testing"
)

func TestDispatchReadPK(t *testing.T) {
	for i := 0; i < 100; i++ {
		expected := float64(200)
		resultPkg := DispatchReadPK(expected, 11, true)
		var result float64
		for _, v := range resultPkg {
			//log.Debug("index %d, got %.2f", i, v)
			result += v
		}
		// 保留六位
		result, _ = strconv.ParseFloat(fmt.Sprintf("%.6f", result), 64)
		if expected != result {
			log.Fatal("expected %f, but %f", expected, result)
		}
	}
}
