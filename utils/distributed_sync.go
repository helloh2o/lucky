package utils

import (
	"context"
	"github.com/helloh2o/lucky/cache"
	"github.com/helloh2o/lucky/log"
	"sync"
	"time"
)

// 等待者
type Waiter struct {
	Id string
	sync.Mutex
	channels map[string]chan struct{}
	OnNotify func(string)
	wk       map[string]int
}

var waiter = Waiter{channels: make(map[string]chan struct{}), wk: make(map[string]int)}

// RDLockOpWait redis 等待分布式锁，直到获取锁
func RDLockOpWait(operation string) func() {
Try:
	done, ok, wait := do(operation, time.Hour)
	if !ok {
		<-wait
		goto Try
	}
	return done
}

// RDLockOpWaitTimeout 自定义超时操作
func RDLockOpWaitTimeout(operation string, timeout time.Duration) func() {
Try:
	done, ok, wait := do(operation, timeout)
	if !ok {
		<-wait
		goto Try
	}
	return done
}

// RDLockOpWithContext redis 上下文获取锁
func RDLockOpWithContext(ctx context.Context, operation string) (func(), bool) {
Try:
	done, ok, wait := do(operation, time.Hour)
	if !ok {
		select {
		case <-ctx.Done():
			return func() {}, true
		case <-wait:
			goto Try
		}
	}
	return done, false
}

// RDLockOp redis 分布式锁, 足够的操作逻辑时间 release释放，got 是否获取锁,wait在线等待，直到获取锁
func RDLockOp(operation string) (release func(), got bool, wait chan struct{}) {
	// 默认10秒放锁
	defaultExpired := time.Second * 10
	return do(operation, defaultExpired)
}

// RDLockOpTimeout redis 分布式锁， 时间不够可能引发并发同步问题 release释放，got 是否获取锁,wait在线等待，直到获取锁
func RDLockOpTimeout(operation string, timeout time.Duration) (release func(), got bool, wait chan struct{}) {
	return do(operation, timeout)
}

// 返回解锁回调和释放获取到锁
func do(key string, expired time.Duration) (func(), bool, chan struct{}) {
	var wc chan struct{}
	var ok bool
	waiter.Lock()
	defer waiter.Unlock()
	if cache.RedisC.SetNX(context.Background(), key, 1, expired).Val() {
		if wc, ok = waiter.channels[key]; !ok {
			wc = make(chan struct{}, 1)
			waiter.channels[key] = wc
		}
		// write wc channel at redis expired nx
		time.AfterFunc(expired, func() {
			waiter.Lock()
			defer waiter.Unlock()
			select {
			case waiter.channels[key] <- struct{}{}:
				if waiter.OnNotify != nil {
					waiter.OnNotify(key)
				}
			default:
			}
		})
		// release resource
		release := func() {
			waiter.Lock()
			defer waiter.Unlock()
			// 释放通知
			if waiter.OnNotify != nil {
				waiter.OnNotify(key)
			}
			// waiter channel is existed
			if _, existed := waiter.channels[key]; !existed {
				return
			}
			cache.RedisC.Del(context.Background(), key)
			if waiter.wk[key] > 0 {
				// must bigger than 1
				waiter.wk[key] -= 1
			}
			log.Debug("key:%s, waiter size:%d", key, waiter.wk[key])
			select {
			case waiter.channels[key] <- struct{}{}:
				// del on all req wait done
				if waiter.wk[key] == 0 {
					delete(waiter.channels, key)
					delete(waiter.wk, key)
					log.Debug("===> key:%s, all sync lock request wait done. <===", key)
				}
			default:
			}
		}
		return release, true, nil
	} else {
		// 等待着数量
		if _, ok = waiter.wk[key]; ok {
			waiter.wk[key] += 1
		} else {
			// at least 2 op
			waiter.wk[key] = 2
		}
		// 这里可能没有，读取redis ttl
		if wc, ok = waiter.channels[key]; !ok {
			wc = make(chan struct{}, 1)
			waiter.channels[key] = wc
			left := cache.RedisC.TTL(context.Background(), key).Val()
			log.Release("key::%s, ttl:%d s", key, left/time.Second)
			time.AfterFunc(left, func() {
				log.Debug("redis key:%s expired.", key)
				// 时间到，写入等待着
				select {
				case wc <- struct{}{}:
					if waiter.OnNotify != nil {
						waiter.OnNotify(key)
					}
				default:
				}
			})
		}
		return func() {}, false, wc
	}
}
