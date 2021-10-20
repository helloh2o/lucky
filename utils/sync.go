package utils

import (
	"github.com/patrickmn/go-cache"
	"time"
)

var syncObjChan = cache.New(time.Minute, time.Minute*5)

// SyncObjByStr 锁定一个字符串的同步操作
func SyncObjByStr(objKey string) func() {
	_, ok := syncObjChan.Get(objKey)
	if !ok {
		syncObjChan.Set(objKey, make(chan struct{}, 1), time.Minute)
	}
	chObj, _ := syncObjChan.Get(objKey)
	// lock the objKey
	objChan := chObj.(chan struct{})
	objChan <- struct{}{}
	return func() {
		// release the objChan
		<-objChan
	}
}
