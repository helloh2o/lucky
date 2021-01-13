package utils

import (
	"github.com/helloh2o/lucky/log"
	"sync"
	"time"
)

// Limiter package lv default limiter
var Limiter *LimiterMap

func init() {
	Limiter = &LimiterMap{
		data:  make(map[interface{}]int64),
		tiker: time.NewTicker(time.Second * 30),
	}
	go Limiter.Clean()
}

// NewLimiter create limiter map
func NewLimiter(timeout int64) *LimiterMap {
	l := &LimiterMap{
		data:    make(map[interface{}]int64),
		tiker:   time.NewTicker(time.Second * time.Duration(timeout)),
		timeout: timeout,
	}
	return l
}

// LimiterMap for limiter
type LimiterMap struct {
	sync.RWMutex
	data    map[interface{}]int64
	tiker   *time.Ticker
	timeout int64
}

func (l *LimiterMap) Add(key interface{}) {
	l.Lock()
	defer l.Unlock()
	l.data[key] = time.Now().Unix()
	log.Debug("Limiter Add key %v", key)
}

func (l *LimiterMap) UnSafeDel(key interface{}) {
	log.Debug("Limiter UnSafeDel key %v", key)
	delete(l.data, key)
}

func (l *LimiterMap) Del(key interface{}) {
	l.Lock()
	defer l.Unlock()
	delete(l.data, key)
	log.Debug("Limiter Del key %v", key)
}

func (l *LimiterMap) IsLimited(key interface{}, seconds int64) bool {
	l.RLock()
	// read
	v, ok := l.data[key]
	l.RUnlock()
	if !ok {
		// safe write
		l.Add(key)
		return false
	}
	if time.Now().Unix()-v < seconds {
		return true
	}
	// safe delete
	l.Del(key)
	return false
}

// Clean self clean
func (l *LimiterMap) Clean() {
	for {
		<-l.tiker.C
		// read need clean keys
		timeoutKeys := make([]interface{}, 0)
		l.RLock()
		now := time.Now().Unix()
		for k, v := range l.data {
			if now-v >= l.timeout {
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
