package utils

import (
	"github.com/helloh2o/lucky/log"
	"sync"
	"time"
)

var Limiter *LimiterMap

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
	limit int64
}

func (l *LimiterMap) Add(key interface{}, limit int64) {
	l.Lock()
	defer l.Unlock()
	l.data[key] = limitItem{
		time.Now(),
		limit,
	}
	log.Debug("Limiter Add key %v", key)
}

func (l *LimiterMap) UnsafeAdd(key interface{}, limit int64) {
	l.data[key] = limitItem{
		time.Now(),
		limit,
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
		l.UnsafeAdd(key, seconds)
		return false
	}
	if time.Now().Before(v.t.Add(time.Second * time.Duration(seconds))) {
		return true
	} else if time.Now().After(v.t) {
		// repeat
		l.UnsafeAdd(key, v.limit)
	}
	return false
}

// Clean self clean
func (l *LimiterMap) Clean() {
	for {
		<-l.tiker.C
		// read need clean keys
		timeoutKeys := make([]interface{}, 0)
		l.RLock()
		for k, v := range l.data {
			if time.Now().After(v.t.Add(time.Second * time.Duration(v.limit))) {
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
