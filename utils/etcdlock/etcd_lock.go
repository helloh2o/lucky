package etcdlock

import (
	"context"
	"errors"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/helloh2o/lucky/log"
	"sync"
	"time"
)

// EtcdLock 基于ETCD KV的分布式锁
type EtcdLock struct {
	mx       sync.Mutex
	cli      *clientv3.Client
	emptyFun func()
	lockMap  map[*concurrency.Mutex]*concurrency.Session
}

// NewDistributedLock 创建新分布式对象
func NewDistributedLock(c *clientv3.Client) (*EtcdLock, func()) {
	el := &EtcdLock{cli: c, emptyFun: func() {}, lockMap: make(map[*concurrency.Mutex]*concurrency.Session)}
	return el, el.Release
}

// Lock 加锁等待
func (el *EtcdLock) Lock(op string) (func(), error) {
	return el.lock(context.TODO(), op)
}

// LockWithTimeout 加锁等待到超时
func (el *EtcdLock) LockWithTimeout(op string, duration time.Duration) (func(), error) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()
	return el.lock(timeoutCtx, op)
}

func (el *EtcdLock) lock(ctx context.Context, op string) (func(), error) {
	var err error
	var session *concurrency.Session
	if el.cli == nil {
		return el.emptyFun, errors.New("etcd client is nil")
	}
	if session, err = concurrency.NewSession(el.cli); err == nil {
		mtx := concurrency.NewMutex(session, op)
		if err = mtx.Lock(ctx); err == nil {
			el.mx.Lock()
			defer el.mx.Unlock()
			el.lockMap[mtx] = session
			lockKey := mtx.Key()
			log.Debug("add lock key %s", lockKey)
			return func() {
				el.mx.Lock()
				defer el.mx.Unlock()
				_ = mtx.Unlock(context.TODO())
				_ = session.Close()
				delete(el.lockMap, mtx)
				log.Debug("release key %s", lockKey)
			}, nil
		}
	}
	return el.emptyFun, err
}

// Release 清理锁
func (el *EtcdLock) Release() {
	el.mx.Lock()
	defer el.mx.Unlock()
	log.Debug("lock map size:%d", len(el.lockMap))
	for mtx, session := range el.lockMap {
		log.Debug("final release all unlock %s", mtx.Key())
		_ = mtx.Unlock(context.TODO())
		_ = session.Close()
		delete(el.lockMap, mtx)
	}
	_ = el.cli.Close()
}
