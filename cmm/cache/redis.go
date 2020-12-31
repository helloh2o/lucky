package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"net/url"
	"strconv"
	"time"
)

// RedisC is the *redis.Client
var RedisC *redis.Client

// OpenRedis redis://localhost:6379/?pwd=&db=
func OpenRedis(rURL string) (*redis.Client, error) {
	urlInfo, err := url.Parse(rURL)
	if err != nil {
		return nil, err
	}
	pwd := urlInfo.Query().Get("pwd")
	dbName := urlInfo.Query().Get("db")
	dbIndex := 0
	if dbName != "" {
		dbIndex, err = strconv.Atoi(dbName)
		if err != nil {
			return nil, err
		}
	}
	instance := redis.NewClient(&redis.Options{
		Addr:     urlInfo.Host,
		Password: pwd,
		DB:       dbIndex,
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err = instance.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	if RedisC == nil {
		RedisC = instance
	}
	return instance, nil
}
