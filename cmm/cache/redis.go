package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"net/url"
	"strconv"
	"time"
)

var RedisC *redis.Client

// redis://localhost:6379/?pwd=&db=
func OpenRedis(rURL string) error {
	urlInfo, err := url.Parse(rURL)
	if err != nil {
		return err
	}
	pwd := urlInfo.Query().Get("pwd")
	dbName := urlInfo.Query().Get("db")
	dbIndex := 0
	if dbName != "" {
		dbIndex, err = strconv.Atoi(dbName)
		if err != nil {
			return err
		}
	}
	RedisC = redis.NewClient(&redis.Options{
		Addr:     urlInfo.Host,
		Password: pwd,
		DB:       dbIndex,
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err = RedisC.Ping(ctx).Result()
	if err != nil {
		return err
	}
	return nil
}
