package main

import (
	"flag"
	"github.com/mlogclub/simple"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"log"
	"lucky-day/conf"
	"os"
	"time"
)

var configFile = flag.String("config", "./lucky.yaml", "配置文件路径")

func init() {
	flag.Parse()
	// gorm配置
	gormConf := &gorm.Config{}
	// 初始化配置
	config := conf.Init(*configFile)
	// 初始化日志
	if file, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
		logrus.SetOutput(io.MultiWriter(os.Stdout, file))
		if config.ShowSql {
			gormConf.Logger = logger.New(log.New(file, "\r\n", log.LstdFlags), logger.Config{
				SlowThreshold: time.Second,
				Colorful:      true,
				LogLevel:      logger.Info,
			})
		}
	} else {
		logrus.SetOutput(os.Stdout)
		logrus.Error(err)
	}

	// 连接数据库
	if err := simple.OpenDB(config.MySqlUrl, gormConf, 10, 20, nil); err != nil {
		logrus.Error(err)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("Lucky:")
	log.Println("hello, what a lucky day.")

}
