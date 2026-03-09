package utils

import (
	"errors"
	"log"
	"sync"
	"time"
)

var opChannelMap sync.Map

// OpLocked 单机锁一个操作字符串
func OpLocked(key string) (func(), error) {
	// 避免竞态条件
	actual, _ := opChannelMap.LoadOrStore(key, make(chan struct{}, 1))
	kc := actual.(chan struct{})

	select {
	case kc <- struct{}{}:
		// 成功获取锁
		return func() {
			// 释放锁
			select {
			case <-kc:
				// 成功释放
			default:
			}
		}, nil
	default:
		return nil, errors.New("can't acquire lock")
	}
}

// OpLockedWait 单机锁一个操作字符串，等待
func OpLockedWait(key string) func() {
	// 避免竞态条件
	actual, _ := opChannelMap.LoadOrStore(key, make(chan struct{}, 1))
	kc := actual.(chan struct{})
	kc <- struct{}{}
	// 成功获取锁
	return func() {
		// 释放锁
		select {
		case <-kc:
			// 成功释放
		default:
		}
	}
}

// OpLockTimeout 单机锁一个操作字符串可超时
func OpLockTimeout(key string, timeout time.Duration) func() {
	// 避免竞态条件
	actual, _ := opChannelMap.LoadOrStore(key, make(chan struct{}, 1))
	kc := actual.(chan struct{})

	select {
	case kc <- struct{}{}:
		// 成功获取锁
		return func() {
			// 释放锁
			select {
			case <-kc:
				// 成功释放
			default:
			}
		}
	case <-time.After(timeout):
		// 超时未获取到锁
		log.Printf("unexpected key %s lock timeout", key)
		return func() {}
	}
}

// DelOpKey 删除锁定的建
func DelOpKey(key string) {
	opChannelMap.Delete(key)
}
