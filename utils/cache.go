package utils

import (
	"context"
	"github.com/helloh2o/lucky/cache"
)

// 是否存在redis key
func ExistedCache(key string) bool {
	n, _ := cache.RedisC.Exists(context.Background(), key).Result()
	return n == 1
}
