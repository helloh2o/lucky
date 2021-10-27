package utils

import (
	"errors"
	"github.com/google/uuid"
	"github.com/helloh2o/lucky/log"
	"runtime/debug"
	"sync"
	"time"
)

// LazyQueue 排队保存，重复的数据只排一次，时效性一般的情况，场景：降低数据库写压力
type LazyQueue struct {
	sync.RWMutex
	saveQueue chan interface{}
	queued    map[interface{}]string
	call      func(interface{}) error
	qps       int
}

func NewLazyQueue(qps, size int, cf func(interface{}) error) (*LazyQueue, error) {
	if cf == nil {
		errMsg := "nil callback is invalid"
		log.Error(errMsg)
		return nil, errors.New(errMsg)
	}
	if size <= 0 {
		return nil, errors.New("lazy queue size is too small")
	}
	if qps <= 0 {
		return nil, errors.New("lazy queue qps is too small")
	}
	lq := &LazyQueue{
		RWMutex:   sync.RWMutex{},
		saveQueue: make(chan interface{}, size),
		queued:    make(map[interface{}]string),
		qps:       qps,
	}
	lq.call = cf
	return lq, nil
}

// Run 启动LazySave 并返回错误
func (lazy *LazyQueue) Run() {
	defer func() {
		if r := recover(); r != nil {
			log.Error("panic %s", string(debug.Stack()))
		}
	}()
	for {
		select {
		case token := <-lazy.saveQueue:
			lazy.callback(token)
		default:
		}
		sleepMills := 1000 / lazy.qps
		time.Sleep(time.Millisecond * time.Duration(sleepMills))
	}
}

// PushToQueue 将不重要的对象加入保存队列
func (lazy *LazyQueue) PushToQueue(key interface{}) {
	lazy.RLock()
	val, ok := lazy.queued[key]
	lazy.RUnlock()
	if ok {
		// key is in processing or not
		done := SyncObjByStr(val)
		defer done()
		lazy.RLock()
		_, ok := lazy.queued[key]
		if ok {
			lazy.RUnlock()
			return
		}
		lazy.RUnlock()
	}
	select {
	case lazy.saveQueue <- key:
		lazy.Lock()
		defer lazy.Unlock()
		lazy.queued[key] = uuid.New().String()
	default:
		// 队列已满
		log.Error("lazy queue is full %d", len(lazy.saveQueue))
	}
}

// OutOfQueue 解除对象慢保存排队
func (lazy *LazyQueue) OutOfQueue(key interface{}) {
	lazy.Lock()
	defer lazy.Unlock()
	if _, ok := lazy.queued[key]; ok {
		// 删除
		delete(lazy.queued, key)
		log.Debug("delete from queue :: %v", key)
	}
}

// 处理数据
func (lazy *LazyQueue) callback(key interface{}) {
	defer func() {
		lazy.Lock()
		defer lazy.Unlock()
		delete(lazy.queued, key)
	}()
	lazy.RLock()
	val, ok := lazy.queued[key]
	defer lazy.RUnlock()
	if ok {
		// key is in processing
		done := SyncObjByStr(val)
		defer done()
		if err := lazy.call(key); err != nil {
			log.Error("lazy call error:: %s", err.Error())
		}
	}
	log.Debug("queue left size %d", len(lazy.saveQueue))
}
