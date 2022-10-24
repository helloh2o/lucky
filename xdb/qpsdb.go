package xdb

import (
	"github.com/helloh2o/lucky/utils"
	"gorm.io/gorm"
	"time"
)

var (
	tokenLimiter *utils.RateLimiter
)

// InitQpsDB 初始化同步等待QPS DB
func InitQpsDB(limit int, interval time.Duration) {
	if limit > 3000 {
		limit = 3000
	}
	tokenLimiter = utils.New(limit, interval)
}

// QqsDB 获取DB对象，若超时则返回nil
func QqsDB() *gorm.DB {
	tokenLimiter.Wait()
	return DB()
}
