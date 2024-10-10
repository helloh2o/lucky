package etcdlock

import (
	"log"
	"sync"
	"testing"
	"time"
)

const testOpKey = "123456"

func TestEtcdLock_Lock(t *testing.T) {
	release := InitDefault("localhost:2379")
	if done, err := D().Lock(testOpKey); err == nil {
		log.Println("s1 got lock")
		time.AfterFunc(time.Second*3, func() {
			log.Println("s1 release lock")
			done()
		})
	}
	if done, err := D().LockWithTimeout(testOpKey, time.Second*5); err != nil {
		log.Printf("s2 got lock error %v", err)
	} else {
		log.Println("s2 got lock")
		done()
	}
	_, _ = D().LockWithTimeout(testOpKey, time.Second*5)
	release()
}

func TestEtcdLock_Lock2(t *testing.T) {
	release := InitDefault("localhost:2379")
	D().Lock(testOpKey)
	_ = release
	log.Println("s1 release lock")
}

// 测试并发获取
func TestEtcdLock_Lock3(t *testing.T) {
	release := InitDefault("localhost:2379")
	defer release()
	wg := sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			if done, err := D().LockWithTimeout(testOpKey, time.Second*5); err != nil {
				wg.Done()
			} else {
				log.Printf("s%d get lock", idx)
				time.AfterFunc(time.Second*time.Duration(idx), func() {
					log.Printf("s%d release lock", idx)
					done()
					wg.Done()
				})
			}
		}(i)
	}
	wg.Wait()
}
