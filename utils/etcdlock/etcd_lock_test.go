package etcdlock

import (
	"github.com/coreos/etcd/clientv3"
	"log"
	"testing"
	"time"
)

const testOpKey = "123456"

func TestEtcdLock_Lock(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	l, release := NewDistributedLock(cli)
	if done, err := l.Lock(testOpKey); err == nil {
		log.Println("s1 got lock")
		time.AfterFunc(time.Second*3, func() {
			done()
			log.Println("s1 release lock")
		})
	}
	if done, err := l.LockWithTimeout(testOpKey, time.Second*5); err != nil {
		log.Printf("s2 got lock error %v", err)
	} else {
		log.Println("s2 got lock")
		done()
	}
	l.LockWithTimeout(testOpKey, time.Second*5)
	release()
}

func TestEtcdLock_Lock2(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	l, release := NewDistributedLock(cli)
	l.Lock(testOpKey)
	_ = release
	log.Println("s1 release lock")
}
