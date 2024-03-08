package xdb

import (
	"github.com/helloh2o/lucky/log"
	"github.com/helloh2o/lucky/utils"
	"gorm.io/gorm"
	"time"
)

var (
	tokenLimiter *utils.RateLimiter
	mainDB       *gorm.DB
)

// SetMainDB 设置主DB
func SetMainDB(main *gorm.DB, DSN string) {
	if main == nil {
		panic("mainDB is nil")
	}
	log.Release("mainDB is set, DSN :%s", DSN)
	mainDB = main
}

// InitQpsDB 初始化同步等待QPS DB
func InitQpsDB(limit int, interval time.Duration) {
	if limit > 5000 {
		log.Error("db qps limit %d maybe too big", limit)
	}
	tokenLimiter = utils.New(limit, interval)
}

// QpsDB 获取DB对象，若超时则返回nil
func QpsDB() *gorm.DB {
	if mainDB == nil {
		panic("mainDB is nil")
	}
	tokenLimiter.Wait()
	return mainDB
}
