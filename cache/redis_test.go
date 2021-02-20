package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/helloh2o/lucky/log"
	"testing"
	"time"
)

func TestOpenRedis(t *testing.T) {
	if _, err := OpenRedis("redis://localhost:6379/?pwd=&db="); err != nil {
		panic(err)
	}
	_, err := RedisC.Do(context.Background(), "SET", "name", "luck-name").Result()
	if err != nil {
		panic(err)
	}
	if _, err = RedisC.Expire(context.Background(), "name", time.Second*3).Result(); err != nil {
		panic(err)
	}
	time.Sleep(time.Second * 3)
	val, err := RedisC.Get(context.Background(), "name").Result()
	if err != nil && err != redis.Nil {
		panic(err)
	}
	log.Debug(val)
}
