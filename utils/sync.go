package utils

import (
	"github.com/patrickmn/go-cache"
	"sync"
	"time"
)

var (
	lk          sync.Mutex
	syncObjChan = cache.New(time.Minute, time.Minute*5)
	tenMinutes  = time.Minute * 10
)

// SyncObjByStr 锁定一个字符串的同步操作
func SyncObjByStr(objKey string) func() {
	// lock the objKey
	objChan := getChannel(objKey, tenMinutes)
	objChan <- struct{}{}
	return func() {
		// release the objChan
		<-objChan
	}
}

func getChannel(objKey string, dur time.Duration) chan struct{} {
	lk.Lock()
	defer lk.Unlock()
	chObj, ok := syncObjChan.Get(objKey)
	if !ok {
		chObj = make(chan struct{}, 1)
	}
	syncObjChan.Set(objKey, chObj, dur)
	return chObj.(chan struct{})
}

// SyncStrWithTimeout 锁定一个字符串的同步操作,允许过期
func SyncStrWithTimeout(objKey string, duration time.Duration) func() {
	// lock the objKey
	objChan := getChannel(objKey, duration)
	objChan <- struct{}{}
	return func() {
		// release the objChan
		<-objChan
	}
}
