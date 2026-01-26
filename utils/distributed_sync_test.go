package utils

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"os"
	"os/signal"
	"saas/log"
	"sync"
	"syscall"
	"testing"
	"time"
)

var (
	rdb   *redis.Client
	opKey = "github.com"
)

func Init(node string) func() {
	rdb = redis.NewClient(&redis.Options{Addr: fmt.Sprintf("%s:%d", "192.168.0.34", 6379)})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	return InitDsbLocker(node, rdb)
}

func TestDsbLockerContext(t *testing.T) {
	done := Init("K1")
	defer done()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	release, timeout := DsbLock().RDLockWithContext(ctx, opKey)
	if timeout {
		log.Info("can't got lock with timeout")
	} else {
		defer release()
		log.Info("ctx op got lock")
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

func TestKeepLock(t *testing.T) {
	done := Init("K0")
	defer done()
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if release, timeout := DsbLock().RDLockWithContextExp(timeoutCtx, opKey, time.Minute*10); timeout {
		panic("can't got lock with timeout")
	} else {
		_ = release
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

func TestKeepLock2(t *testing.T) {
	done := Init("K0")
	defer done()
	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if release, timeout := DsbLock().RDLockWithContextExp(timeoutCtx, opKey, time.Minute*10); timeout {
		panic("can't got lock with timeout")
	} else {
		_ = release
	}
	select {}
}

func TestDsbLockerWait(t *testing.T) {
	done := Init("G0")
	defer done()
	for i := 0; i < 10; i++ {
		release := DsbLock().RDLockWait(opKey)
		log.Info("wait op got lock")
		time.Sleep(time.Second * 5)
		release()
	}
}

func TestDsbLockerWaitExpG1(t *testing.T) {
	done := Init("G1")
	defer done()
	wg := sync.WaitGroup{}
	// 模拟G1组 100个人不停抢锁
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			idx := index
			release := DsbLock().RDLockWaitHour(opKey)
			log.Info("NO.%d wait ex op got lock\n", idx)
			time.Sleep(time.Second)
			release()
		}(i + 1)
	}
	wg.Wait()
}

func TestDsbLockerWaitExpG2(t *testing.T) {
	done := Init("")
	defer done()
	wg := sync.WaitGroup{}
	for i := 0; i < 2000; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			release := DsbLock().RDLockWaitHour(opKey)
			log.Info("G2 NO.%d wait ex op got lock\n", index)
			time.Sleep(time.Second)
			release()
		}(i + 1)
	}
	wg.Wait()
}

func TestDsbLockerWaitExpG3(t *testing.T) {
	done := Init("G3")
	defer done()
	wg := sync.WaitGroup{}
	for i := 0; i < 3000; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			release := DsbLock().RDLockWaitHour(opKey)
			log.Info("G3 NO.%d wait ex op got lock\n", index)
			time.Sleep(time.Second)
			release()
		}(i + 1)
	}
	wg.Wait()
}

func TestDsbLockerWaitExpG4(t *testing.T) {
	done := Init("G4")
	defer done()
	release := DsbLock().RDLockWait(opKey)
	log.Info("G4  wait ex op got lock")
	time.Sleep(time.Minute)
	release()
}
