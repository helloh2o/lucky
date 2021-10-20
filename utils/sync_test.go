package utils

import (
	"log"
	"testing"
	"time"
)

func TestSyncObjByStr(t *testing.T) {
	outerOk := SyncObjByStr("ok")
	for i := 0; i < 10; i++ {
		go func(index int) {
			innerOk := SyncObjByStr("ok")
			defer func(index int) {
				innerOk()
				log.Printf("release inner ok for i=%d", index)
			}(index)
		}(i)
		time.Sleep(time.Second)
	}
	time.Sleep(time.Second * 10)
	log.Printf("release outer ok")
	outerOk()
	select {}
}
