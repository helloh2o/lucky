package etcdlock

import (
	"log"
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
