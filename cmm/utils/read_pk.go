package utils

import (
	"fmt"
	"github.com/helloh2o/lucky/log"
	"golang.org/x/exp/rand"
	"math"
	"strconv"
	"time"
)

// 保留两位小数
var r = rand.New(rand.NewSource(uint64(time.Now().UnixNano())))

// DispatchReadPK  random || @average money/amount
func DispatchReadPK(money float64, amount int, average bool) (pks []float64) {
	if money == 0 || amount == 0 {
		log.Error("money & amount must bigger than 0.")
		return
	}
	// 概率列表
	probList := make([]float64, amount)
	AllProb := float64(0)
	for i := 0; i < amount; i++ {
		if average {
			probList[i] = 1
			AllProb += 1
		} else {
			ni := r.Intn(100) + 1
			AllProb += float64(ni)
			probList[i] = float64(ni)
		}
	}
	var allocated float64
	// read pk
	pks = make([]float64, amount)
	for i := 0; i < amount; i++ {
		per := (probList[i] / AllProb) * money * 100
		per = math.Floor(per) / 100
		pks[i] = per
		allocated += per * 100
	}
	if allocated/100 != money {
		// 剩下的小部分，加给某一个人
		lucky := r.Intn(amount)
		left := money - allocated/100
		//log.Debug("allocated=%f money=%f left part %f", allocated/ 100, money, left)
		pks[lucky], _ = strconv.ParseFloat(fmt.Sprintf("%.2f", pks[lucky]+left), 64)
	}
	log.Debug("red pkgs %+v", pks)
	return
}
