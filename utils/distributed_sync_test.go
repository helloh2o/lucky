package utils

import (
	"github.com/helloh2o/lucky/cache"
	"github.com/helloh2o/lucky/log"
	"sync"
	"testing"
	"time"
)

func init() {
	// 初始化Redis
	if _, err := cache.OpenRedis("redis://localhost:6379/?pwd=&db=0"); err != nil {
		panic(err)
	}
}
func TestRDLockOp(t *testing.T) {
	op := "select_box"
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(owner int) {
			defer wg.Done()
			if release, ok, wait := RDLockOp(op); ok {
				log.Release("owner::%d get lock", owner)
				<-time.Tick(time.Second * time.Duration(RandInterval(1, 5)))
				defer release()
			} else {
				<-wait
				if release, ok, _ := RDLockOp(op); ok {
					defer release()
					log.Release("owner::%d, wait release ok.", owner)
				}
			}
		}(i)
	}
	wg.Wait()
}
