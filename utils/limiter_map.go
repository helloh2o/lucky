package utils

import (
	"github.com/helloh2o/lucky/log"
	"sync"
	"sync/atomic"
	"time"
)

var Limiter *LimiterMap

const tiktok = 1

func init() {
	Limiter = &LimiterMap{
		data:  make(map[interface{}]limitItem),
		tiker: time.NewTicker(time.Second * 30),
	}
	go Limiter.Clean()
}

type LimiterMap struct {
	sync.RWMutex
	data  map[interface{}]limitItem
	tiker *time.Ticker
}

type limitItem struct {
	t     time.Time
	limit time.Duration
	times int64
}

func (l *LimiterMap) Add(key interface{}, duration time.Duration) {
	l.Lock()
	defer l.Unlock()
	l.data[key] = limitItem{
		time.Now(),
		duration,
		tiktok,
	}
	log.Debug("Limiter Add key %v", key)
}

func (l *LimiterMap) UnsafeAdd(key interface{}, duration time.Duration) {
	l.data[key] = limitItem{
		time.Now(),
		duration,
		tiktok,
	}
	log.Debug("Limiter Add key %v", key)
}

func (l *LimiterMap) Del(key interface{}) {
	l.Lock()
	defer l.Unlock()
	delete(l.data, key)
	log.Debug("Limiter Del key %v", key)
}

func (l *LimiterMap) UnSafeDel(key interface{}) {
	log.Debug("Limiter UnSafeDel key %v", key)
	delete(l.data, key)
}

func (l *LimiterMap) IsLimited(key interface{}, seconds int64) bool {
	l.Lock()
	defer l.Unlock()
	// read
	v, ok := l.data[key]
	if !ok {
		// safe write
		l.UnsafeAdd(key, time.Second*time.Duration(seconds))
		return false
	}
	atomic.AddInt64(&v.times, tiktok)
	if time.Now().Before(v.t.Add(time.Second * time.Duration(seconds))) {
		l.data[key] = v
		return true
	} else {
		// repeat
		l.UnsafeAdd(key, v.limit)
	}
	return false
}

func (l *LimiterMap) IsV2Limited(key interface{}, duration time.Duration, max int64) (bool, int64) {
	l.Lock()
	defer l.Unlock()
	// read
	v, ok := l.data[key]
	if !ok {
		// safe write
		l.UnsafeAdd(key, duration)
		return false, tiktok
	}
	atomic.AddInt64(&v.times, tiktok)
	l.data[key] = v
	if time.Now().Before(v.t.Add(duration)) {
		log.Release("v.times:%d , the max:%d", v.times, max)
		if v.times > max {
			return true, v.times
		}
		return false, v.times
	} else {
		// repeat
		l.UnsafeAdd(key, v.limit)
	}
	return false, tiktok
}

// Clean self clean
func (l *LimiterMap) Clean() {
	for {
		<-l.tiker.C
		// read need clean keys
		timeoutKeys := make([]interface{}, 0)
		l.RLock()
		for k, v := range l.data {
			if time.Now().After(v.t.Add(v.limit)) {
				timeoutKeys = append(timeoutKeys, k)
			}
		}
		l.RUnlock()
		// write data
		l.Lock()
		for _, k := range timeoutKeys {
			l.UnSafeDel(k)
		}
		l.Unlock()
	}
}
