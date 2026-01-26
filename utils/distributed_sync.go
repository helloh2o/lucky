package utils

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"io"
	"os"
	"saas/log"
	"strings"
	"sync"
	"time"
)

const (
	OnlineNodes        = "dlock:online" //在线节点
	NodeIdKey          = "dlock:nodeId" //自增ID
	NodeFileDir        = "./runtime"
	NodeIdFile         = "./runtime/node_locker_id_"
	NotifyEventChannel = "SET_NX:notify_unlock_event"
)

// RNodeLock 基于Redis的分布式锁
type RNodeLock struct {
	// 节点ID
	Nid string
	sync.Mutex
	channels   map[string]*Waiter
	rdb        redis.UniversalClient
	SelfNotify func(string) // 自定义解锁通知
}

type Waiter struct {
	N         int64
	WC        chan struct{}
	Expired   time.Duration
	ExpiredAt time.Time
	Timer     *time.Timer
}

var defaultLocker = &RNodeLock{channels: make(map[string]*Waiter)}

func DsbLock() *RNodeLock {
	return defaultLocker
}

// InitDsbLocker 初始化默认分布式锁节点
// @nodeId 节点ID
// @rdb Redis 客户端
// @notifies 解锁通知方法
func InitDsbLocker(nodeId string, rdb redis.UniversalClient, notifies ...func(string)) func() {
	newNode := true
	// 优先读取存储的节点ID，如果不存在则生产新节点ID
	nodeFile := NodeIdFile + nodeId
	if f, err := os.Open(nodeFile); err == nil {
		// 读取上次使用的node_id
		if idData, err := io.ReadAll(f); err == nil {
			nodeId = string(idData)
			newNode = false
		} else {
			log.Error("read node id file error:", err)
		}
	}
	// 是否已获得节点名
	if nodeId == "" {
		nodeId = fmt.Sprintf("locker#%d", rdb.IncrBy(context.Background(), NodeIdKey, 1).Val())
	}
	if newNode {
		if _, err := os.Stat(NodeFileDir); err != nil {
			// 创建目录
			if err = os.Mkdir(NodeFileDir, os.ModePerm); err != nil {
				log.Fatalf("create dir error:%v", err)
			}
		}
		// 保存新节点，只读文件
		if err := os.WriteFile(nodeFile, []byte(nodeId), os.ModePerm); err != nil {
			log.Fatalf("write node id file error:%v", err)
		}
	}
	// 节点上线
	if rdb.SAdd(context.Background(), OnlineNodes, nodeId).Val() == 0 {
		// 处理节点非正常退出的情况
		if strings.Contains(nodeId, "locker#") {
			// 自动生成ID
			log.Warn("WARN: rdb distributed lock node <%s> already online.", nodeId)
		} else {
			// 用户自定义的ID可能重复，必须正常退出。如果自定义ID重复，又被异常关闭有极小可能造成锁被异常释放
			log.Fatalf("ERROR: rdb distributed lock node <%s> already online.", nodeId)
		}
	}
	defaultLocker.Nid = nodeId
	defaultLocker.rdb = rdb
	// 是否存在异常退出未释放的锁
	ctx := context.Background()
	// 是否自定义处理通知
	if len(notifies) > 0 {
		defaultLocker.SelfNotify = notifies[0]
	} else {
		// 默认Redis订阅其它节点释放信息
		subscribe := func() {
			// 订阅解锁频道
			psb := rdb.Subscribe(ctx, NotifyEventChannel)
			// 确认订阅成功
			if _, err := psb.Receive(ctx); err != nil {
				log.Error("subscribe err:", err)
				return
			}
			// 通过通道接收消息
			ch := psb.Channel()
			for msg := range ch {
				//log.Info("收到消息: 频道=%s 内容=%s\n", msg.Channel, msg.Payload)
				defaultLocker.HandleMsg(msg.Payload)
			}
		}
		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Error("subscribe panic err:%v", err)
				}
			}()
			// subscribe forever
			for {
				log.Info("node:%s subscribe start.", nodeId)
				subscribe()
				log.Warn("node:%s subscribe end.", nodeId)
				time.Sleep(time.Second)
			}
		}()
	}
	// 重置当前节点的可能异常退出没有释放的锁
	if keys, err := rdb.HGetAll(ctx, defaultLocker.Nid).Result(); err == nil {
		for key, _ := range keys {
			go func(last string) {
				defer func() {
					if errRelease := recover(); err != nil {
						log.Info("release old key panic err:%v", errRelease)
					}
				}()
				// delete hold
				rdb.HDel(ctx, nodeId, last)
				// expire key and notify all
				expired := time.Second * 10
				rdb.Expire(ctx, last, expired)
				time.Sleep(expired)
				if defaultLocker.SelfNotify != nil {
					defaultLocker.SelfNotify(key)
				} else {
					defaultLocker.Pub(key)
				}
				log.Info("node:%s exception key:%s reset expired.", nodeId, last)
			}(key)
		}
	}
	go defaultLocker.WatchKeys()
	return defaultLocker.Release
}

func (locker *RNodeLock) Pub(key string) {
	if err := locker.rdb.Publish(context.Background(), NotifyEventChannel, key).Err(); err != nil {
		log.Error("pub unlock key err:%v \n", err)
	} else {
		//log.Info("pub unlock key:%s ok \n", key)
	}
}

func (locker *RNodeLock) HandleMsg(msg string) {
	locker.Lock()
	defer locker.Unlock()
	if wt, ok := locker.channels[msg]; ok {
		select {
		case wt.WC <- struct{}{}:
		default:
		}
		//log.Info("notify key:%s by node:%s", key, sender)
	}
}

func (locker *RNodeLock) Sub() {

}

// RDLockWait redis 等待分布式锁，直到获取锁，锁不过期 (一般情况下不推荐使用此方法)
func (locker *RNodeLock) RDLockWait(operation string) func() {
	wtt := 0
WAIT:
	done, ok, wait := locker.do(operation, 0, wtt)
	if !ok {
		<-wait
		wtt++
		goto WAIT
	}
	return done
}

// RDLockWaitHasExp redis 等待分布式锁，直到获取锁, 锁自定义过期时间（确保自己的逻辑时间足够）
func (locker *RNodeLock) RDLockWaitHasExp(operation string, lockExp time.Duration) func() {
	wtt := 0
WAIT:
	done, ok, wait := locker.do(operation, lockExp, wtt)
	if !ok {
		<-wait
		wtt++
		goto WAIT
	}
	return done
}

// RDLockAuto 默认15分钟，自动续期
func (locker *RNodeLock) RDLockAuto(operation string) func() {
	return locker.RDLockWaitHasExp(operation, time.Minute*15)
}

// RDLockWaitHour redis 等待分布式锁，直到获取锁，锁一小时过期 （推荐使用，一小时足够完成逻辑执行）
func (locker *RNodeLock) RDLockWaitHour(operation string) func() {
	return locker.RDLockWaitHasExp(operation, time.Hour)
}

// RDLockAsync 获取锁，不阻塞只返回结果，ok的情况，业务执行defer done()释放锁
func (locker *RNodeLock) RDLockAsync(operation string, lockExp time.Duration) (func(), bool) {
	done, ok, _ := locker.do(operation, lockExp, 0)
	return done, ok
}

// RDLockWithContextExp redis 上下文获取锁， 超时上下文可
func (locker *RNodeLock) RDLockWithContextExp(ctx context.Context, operation string, exp time.Duration) (func(), bool) {
	wtt := 0
WAIT:
	done, ok, wait := locker.do(operation, exp, wtt)
	if !ok {
		select {
		case <-ctx.Done():
			return func() {}, true
		case <-wait:
			wtt++
			goto WAIT
		}
	}
	return done, false
}
func (locker *RNodeLock) RDLockWithContext(ctx context.Context, operation string) (func(), bool) {
	return locker.RDLockWithContextExp(ctx, operation, 0)
}

// 返回解锁回调和释放获取到锁
func (locker *RNodeLock) do(key string, expired time.Duration, wtt int) (func(), bool, chan struct{}) {
	if locker.rdb == nil {
		panic("rdb is nil,please call InitDLocker first")
	}
	var wt *Waiter
	var ok bool
	locker.Lock()
	defer locker.Unlock()
	// 抢锁
	reTry := func(try *Waiter) {
		select {
		// 通知其他等待着可以抢锁了
		case try.WC <- struct{}{}:
		default:
		}
	}
	if locker.rdb.SetNX(context.Background(), key, "1", expired).Val() {
		//log.Info("set nx for key: %s", key)
		// 进程channel
		if wt, ok = locker.channels[key]; !ok {
			wt = &Waiter{WC: make(chan struct{}), N: 1}
			locker.channels[key] = wt
		}
		wt.Expired = expired
		wt.ExpiredAt = time.Now().Add(expired)
		// write wc channel at redis expired nx
		if expired > 0 {
			// 重置timer
			wt.Timer = time.AfterFunc(expired, func() { reTry(wt) })
		}
		// 抢到锁存储到Redis
		locker.rdb.HSet(context.Background(), locker.Nid, key, 1)
		// release resource
		release := func() {
			locker.Lock()
			defer locker.Unlock()
			if _, ok = locker.channels[key]; !ok {
				// channel had released
				return
			}
			// del nx
			locker.rdb.Del(context.Background(), key)
			locker.rdb.HDel(context.Background(), locker.Nid, key)
			wt = locker.channels[key]
			// 释放通知
			if locker.SelfNotify != nil {
				locker.SelfNotify(key)
			}
			wt.N--
			// delete channel
			if wt.N == 0 {
				delete(locker.channels, key)
			}
			locker.Pub(key)
			//log.Info("set nx key:%s, released. waiters left:%d\n", key, wt.N)
			// 亲儿子优先通知
			reTry(wt)
			// 锁持有者主动释放了
			if wt.Timer != nil {
				// 终止timer
				wt.Timer.Stop()
				wt.Timer = nil
			}
		}
		return release, true, nil
	} else {
		// 这里可能没有，读取redis ttl
		if wt, ok = locker.channels[key]; !ok {
			wt = &Waiter{WC: make(chan struct{})}
			locker.channels[key] = wt
			ttl := locker.rdb.TTL(context.Background(), key).Val()
			if ttl > 0 {
				// 有过期时间
				wt.Expired = ttl
				wt.ExpiredAt = time.Now().Add(wt.Expired)
				//log.Info("key::%s, ttl:%d s \n", key, ttl/time.Second)
				wt.Timer = time.AfterFunc(ttl, func() { reTry(wt) })
			}
		}
		// 等待者++
		if wtt == 0 {
			wt.N++
		}
		return func() {}, false, wt.WC
	}
}

func (locker *RNodeLock) Release() {
	locker.Lock()
	defer locker.Unlock()
	if keys, err := locker.rdb.HGetAll(context.Background(), defaultLocker.Nid).Result(); err == nil {
		log.Info("clean all hold keys %d", len(keys))
		// 程序退出释放所有持有的锁
		for key, _ := range keys {
			// 删除节点持有的锁，并通知其它节点
			locker.rdb.Del(context.Background(), key)
			locker.rdb.HDel(context.Background(), locker.Nid, key)
			if locker.SelfNotify != nil {
				locker.SelfNotify(key)
			} else {
				locker.Pub(key)
			}
		}
	}
	// 删掉节点在线状态
	locker.rdb.SRem(context.Background(), OnlineNodes, locker.Nid)
}

func (locker *RNodeLock) WatchKeys() {
	// 每一分钟观察一次
	tk := time.NewTicker(time.Minute)
	for {
		<-tk.C
		SafeCall(func() {
			locker.Lock()
			defer locker.Unlock()
			if keys, err := locker.rdb.HGetAll(context.Background(), defaultLocker.Nid).Result(); err == nil {
				log.Info("watch hold keys %d", len(keys))
				// 程序退出释放所有持有的锁
				future := time.Now().Add(time.Minute * 10)
				for key, _ := range keys {
					if wt, ok := locker.channels[key]; ok {
						// 不过期的KEY
						if wt.Expired == 0 {
							continue
						}
						// 给10分钟后要过期的KEY，续签一个用户定义的过期周期
						if wt.ExpiredAt.Before(future) {
							if err := locker.rdb.Expire(context.Background(), key, wt.Expired).Err(); err == nil {
								// 重置waiter过期
								wt.ExpiredAt = time.Now().Add(wt.Expired)
								log.Debug("keep lock for key:%s", key)
							}
						}
					}
				}
			}
		})
	}
}
