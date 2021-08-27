package utils

import (
	"github.com/helloh2o/lucky/log"
	"time"
)

// RankPool 排序池
type RkPool struct {
	name string
	// 上次排序的结果
	rkSorted []RankItem
	// 需要排序的数据
	rkMap map[interface{}]RankItem
	// 排序周期
	rkCycle *time.Ticker
	// idle 空闲
	idle chan struct{}
	// 最大排序条数
	maxSize int
	// 调用者自己规则排序
	selfSort func([]RankItem) []RankItem
}

// RankItem 排序项目
type RankItem struct {
	// 键
	Key interface{}
	// 原始值
	Data interface{}
	// 排名值
	RankVal int64
}

// NewRKPool 创建一个新的排序池
func NewRKPool(name string, size int, rankCycle time.Duration, selfSort func([]RankItem) []RankItem) *RkPool {
	up := RkPool{
		name:     name,
		rkCycle:  time.NewTicker(rankCycle),
		idle:     make(chan struct{}, 1),
		maxSize:  size,
		selfSort: selfSort,
		rkMap:    make(map[interface{}]RankItem),
	}
	return &up
}

// Queue 排队到池
func (rp *RkPool) Queue(item RankItem) {
	rp.idle <- struct{}{}
	defer func() {
		<-rp.idle
		log.Release("rank_pool data size %d", len(rp.rkMap))
	}()
	_, ok := rp.rkMap[item.Key]
	if ok {
		// 如果已经在队列中，则替换
		rp.rkMap[item.Key] = item
	} else {
		if len(rp.rkMap) >= rp.maxSize {
			log.Release("rk queue %d is full.", rp.maxSize)
		} else {
			rp.rkMap[item.Key] = item
			log.Release("rank_pool add new item: %s, val: %d", item.Key, item.RankVal)
		}
	}
}

// IsInPool 是否在池子中
func (rp *RkPool) IsInPool(item RankItem) bool {
	rp.idle <- struct{}{}
	defer func() {
		<-rp.idle
	}()
	_, ok := rp.rkMap[item.Key]
	return ok
}

// GetRankData 获取已排序数据
func (rp *RkPool) GetRankData(from, to int) (data []RankItem) {
	rp.idle <- struct{}{}
	defer func() {
		<-rp.idle
	}()
	if from > to || from < 1 || len(rp.rkSorted) == 0 {
		return
	}
	if from >= len(rp.rkSorted) {
		from = len(rp.rkSorted)
	}
	if to >= len(rp.rkSorted) {
		to = len(rp.rkSorted)
	}
	return rp.rkSorted[from-1 : to-1]
}

// GetNO1 获取第一个用户
func (rp *RkPool) GetNO1() *RankItem {
	rp.idle <- struct{}{}
	defer func() {
		<-rp.idle
	}()
	if len(rp.rkSorted) > 0 {
		return &rp.rkSorted[0]
	}
	return nil
}

// Remove from pool
func (rp *RkPool) Remove(key interface{}) {
	rp.idle <- struct{}{}
	defer func() {
		<-rp.idle
	}()
	// 删除需要排序的数据
	delete(rp.rkMap, key)
	// 从已排序中删除
	for index, item := range rp.rkSorted {
		if item.Key == key {
			rp.rkSorted = append(rp.rkSorted[:index], rp.rkSorted[index+1:]...)
		}
	}
}

// Serve 开启间隙自动排序
func (rp *RkPool) Serve() {
	for {
		<-rp.rkCycle.C
		// do rank by rp val
		go rp.Rank()
	}
}

// Clean 清空排序池
func (rp *RkPool) Clean() {
	rp.idle <- struct{}{}
	defer func() {
		<-rp.idle
	}()
	if len(rp.rkSorted) > 0 {
		rp.rkSorted = rp.rkSorted[:0]
	}
	rp.rkMap = make(map[interface{}]RankItem)
}

// 进行排序
func (rp *RkPool) Rank() {
	rp.idle <- struct{}{}
	defer func() {
		<-rp.idle
	}()
	var needSortItems []RankItem
	for _, item := range rp.rkMap {
		if val, has := rp.rkMap[item.Key]; has {
			needSortItems = append(needSortItems, val)
		}
	}
	// 按照item rank Val 排序
	for i := 0; i < len(needSortItems)-1; i++ {
		for j := i + 1; j < len(needSortItems); j++ {
			// 从大到小
			if needSortItems[i].RankVal < needSortItems[j].RankVal {
				needSortItems[i], needSortItems[j] = needSortItems[j], needSortItems[i]
			}
		}
	}
	rp.rkSorted = needSortItems
	// 自定义排序
	if rp.selfSort != nil && len(needSortItems) > 1 {
		// 外部规则排序
		rp.rkSorted = rp.selfSort(needSortItems)
	}
	/*for rk, item := range rp.rkSorted {
		log.Debug("rank:: %d, rank value:%d, item:: %v", rk+1, item.RankVal, item.Key)
	}*/
	if len(rp.rkSorted) > 0 {
		size := len(rp.rkSorted)
		log.Release("rank_pool:: %s, sorted items %d, max:%d min:%d", rp.name, size, rp.rkSorted[0].RankVal, rp.rkSorted[len(rp.rkSorted)-1].RankVal)
	}
}
