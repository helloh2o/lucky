package utils

import (
	"fmt"
	"golang.org/x/exp/rand"
	"lucky/log"
	"math"
	"strconv"
	"time"
)

// 红包分发、 参考文章 https://blog.csdn.net/Evrse/article/details/110144412?utm_medium=distribute.pc_feed.none-task-blog-personrec_tag-9.nonecase&depth_1-utm_source=distribute.pc_feed.none-task-blog-personrec_tag-9.nonecase&request_id=5fc021d95b578e08897d84cf
// 保留两位小数
func DispatchReadPK(money float64, amount int) (pks []float64) {
	r := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
	if money == 0 || amount == 0 {
		log.Error("money & amount must bigger than 0.")
		return
	}
	// 概率列表
	probList := make([]float64, amount)
	AllProb := float64(0)
	for i := 0; i < amount; i++ {
		ni := r.Intn(100) + 1
		AllProb += float64(ni)
		probList[i] = float64(ni)
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
		left := money*100 - allocated
		log.Debug("allocated=%f money=%f left part %f", allocated/100, money, left/100)
		pks[lucky], _ = strconv.ParseFloat(fmt.Sprintf("%.2f", pks[lucky]+left/100), 64)
	}
	log.Debug("red pkgs %+v", pks)
	return
}
