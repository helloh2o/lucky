package initialize

import (
	"github.com/go-redis/redis/v8"
	"github.com/helloh2o/lucky/cache"
	"github.com/helloh2o/lucky/im/server/config"
	"github.com/helloh2o/lucky/log"
)

func InitRDB() {
	cfg := config.Get()
	// 初始化Redis 哨兵模式优先
	if len(cfg.RedisSentinelGroup) > 0 {
		log.Release("==> redis on sentinel model <==")
		if sentinelRDB, err := cache.NewSentinelClientOption(&redis.FailoverOptions{
			MasterName:    cfg.RedisMasterName,
			SentinelAddrs: cfg.RedisSentinelGroup,
			DB:            cfg.RedisSDBIndex,
		}); err != nil {
			panic(err)
		} else {
			cache.RedisC = sentinelRDB
		}
	} else {
		// 初始化单机Redis
		if _, err := cache.OpenRedis(cfg.RedisUrl); err != nil {
			panic(err)
		}
	}
	log.Release("redis db is ready!")
}
