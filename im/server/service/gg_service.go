package main

import (
	"flag"
	"github.com/helloh2o/lucky"
	"github.com/helloh2o/lucky/im/server/config"
	"github.com/helloh2o/lucky/im/server/route"
	"github.com/helloh2o/lucky/log"
	"github.com/helloh2o/lucky/natsq"
	"github.com/helloh2o/lucky/utils"
	"github.com/helloh2o/lucky/utils/etcdlock"
	"github.com/kataras/iris/v12/context"
)

var confPath = flag.String("conf", "./config.yaml", "config file path")

func main() {
	flag.Parse()
	cfg := config.Initialize(*confPath)
	jsp := route.JsonHandler()
	_ = natsq.InitOneClient("chat_msg_ns", cfg.NatsUrl, config.Get().ServerId)
	release := etcdlock.InitDefault(cfg.ETCDClusterList...)
	// 跨越
	lucky.EnableCrossOrigin()
	// API注册
	lucky.Post("/msg/send", func(context *context.Context) {
		body, err := context.GetBody()
		if err != nil {
			log.Error("Read body error %v", err)
			return
		} else {
			err = jsp.OnReceivedPackage(context, body)
			log.Error(err)
		}
	})
	// 处理逻辑注册
	route.InitHandler()
	// 优雅退出 made test on ctrl + C  | killall proc | service restart
	graceDone := make(chan struct{})
	exitDone := make(chan struct{})
	go utils.IrisSVExit(graceDone, release)
	// 账号监听服务
	go func() {
		_ = lucky.Run(cfg.ListenAddr)
		<-graceDone
		exitDone <- struct{}{}
	}()
	<-exitDone
}
