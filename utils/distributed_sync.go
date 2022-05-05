package utils

import (
	"context"
	"github.com/helloh2o/lucky/cache"
	"github.com/helloh2o/lucky/log"
	"sync"
	"time"
)

//==================================== 基于redis的可等待分布式锁 =================================================//

// 等待者
type Waiter struct {
	sync.Mutex
	channnels map[string]chan struct{}
}

var waiter = Waiter{channnels: make(map[string]chan struct{})}

// RDLockOp redis 分布式锁, 足够的操作逻辑时间 release释放，got 是否获取锁,wait在线等待，直到获取锁
func RDLockOp(operation string) (release func(), got bool, wait chan struct{}) {
	// 默认10秒放锁
	defaultExpired := time.Second * 10
	return do(operation, defaultExpired)
}

// RDLockOp redis 分布式锁， 时间不够可能引发并发同步问题 release释放，got 是否获取锁,wait在线等待，直到获取锁
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
		if wc, ok = waiter.channnels[key]; !ok {
			wc = make(chan struct{}, 1)
			waiter.channnels[key] = wc
		}
		return func() {
			waiter.Lock()
			defer waiter.Unlock()
			cache.RedisC.Del(context.Background(), key)
			select {
			case waiter.channnels[key] <- struct{}{}:
			default:
			}
		}, true, nil
	} else {
		// 这里可能没有，读取redis ttl
		if wc, ok = waiter.channnels[key]; !ok {
			wc = make(chan struct{}, 1)
			waiter.channnels[key] = wc
			left := cache.RedisC.TTL(context.Background(), key).Val()
			log.Release("key::%s, ttl:%d s", key, left/time.Second)
			time.AfterFunc(left, func() {
				log.Release("redis key:%s expired.", key)
				// 时间到，写入等待着
				select {
				case wc <- struct{}{}:
				default:
				}
			})
		}
		return func() {}, false, wc
	}
}
