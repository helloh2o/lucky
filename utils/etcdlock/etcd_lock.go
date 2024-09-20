package etcdlock

import (
	"context"
	"errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"log"
	"sync"
	"time"
)

var etcdLock *EtcdLock

// InitDefault 初始化默认锁
func InitDefault(endpoints ...string) func() {
	if endpoints == nil || len(endpoints) == 0 {
		endpoints = []string{"127.0.0.1:2379"}
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:           endpoints,
		PermitWithoutStream: true,
		DialTimeout:         5 * time.Second,
		AutoSyncInterval:    10 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	l, release := NewDistributedLock(cli)
	if l == nil {
		panic("etcd lock not init")
	}
	etcdLock = l
	return release
}

// D 返回默认锁
func D() *EtcdLock {
	if etcdLock == nil {
		panic("etcd lock not init")
	}
	return etcdLock
}

// EtcdLock 基于ETCD KV的分布式锁
type EtcdLock struct {
	mx       sync.Mutex
	Cli      *clientv3.Client
	emptyFun func()
	lockMap  map[*concurrency.Mutex]*concurrency.Session
}

// NewDistributedLock 创建新分布式对象
func NewDistributedLock(c *clientv3.Client) (*EtcdLock, func()) {
	el := &EtcdLock{Cli: c, emptyFun: func() {}, lockMap: make(map[*concurrency.Mutex]*concurrency.Session)}
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
	if el.Cli == nil {
		return el.emptyFun, errors.New("etcd client is nil")
	}
	if session, err = concurrency.NewSession(el.Cli); err == nil {
		mtx := concurrency.NewMutex(session, op)
		if err = mtx.Lock(ctx); err == nil {
			el.mx.Lock()
			defer el.mx.Unlock()
			el.lockMap[mtx] = session
			return func() {
				el.mx.Lock()
				defer el.mx.Unlock()
				_ = mtx.Unlock(context.TODO())
				_ = session.Close()
				delete(el.lockMap, mtx)
				//log.Printf("op done delete key %s", op)
			}, nil
		}
	}
	return el.emptyFun, err
}

// Release 清理锁
func (el *EtcdLock) Release() {
	el.mx.Lock()
	defer el.mx.Unlock()
	log.Printf("lock map size:%d", len(el.lockMap))
	for mtx, session := range el.lockMap {
		log.Printf("final release all unlock %s", mtx.Key())
		_ = mtx.Unlock(context.TODO())
		_ = session.Close()
		delete(el.lockMap, mtx)
	}
}
