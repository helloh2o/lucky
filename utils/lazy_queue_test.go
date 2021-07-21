package utils

import (
	"github.com/helloh2o/lucky/log"
	"testing"
)

func TestNewLazyQueue(t *testing.T) {
	lq, err := NewLazyQueue(10, 500, func(i interface{}) error {
		log.Debug("key item: %v", i)
		return nil
	})
	if err != nil {
		panic(err)
	}
	go lq.Run()
	go func() {
		for i := 1; i < 1000; i++ {
			lq.PushToQueue(i)
		}
	}()
	go func() {
		for i := 1; i < 1000; i++ {
			lq.PushToQueue(i)
		}
	}()
	select {}
}
