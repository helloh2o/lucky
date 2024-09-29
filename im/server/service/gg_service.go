package main

import (
	"flag"
	"github.com/helloh2o/lucky"
	"github.com/helloh2o/lucky/im/server/config"
	initialize2 "github.com/helloh2o/lucky/im/server/initialize"
	"github.com/helloh2o/lucky/im/server/route"
	"github.com/helloh2o/lucky/log"
	"github.com/helloh2o/lucky/natsq"
	"github.com/helloh2o/lucky/utils"
	"github.com/helloh2o/lucky/utils/etcdlock"
	"github.com/kataras/iris/v12/context"
	"sync"
)

var (
	confPath   = flag.String("conf", "./config.yaml", "config file path")
	encryptors sync.Map
)

func main() {
	flag.Parse()
	cfg := config.Initialize(*confPath)
	jsp := route.JsonHandler()
	_ = natsq.InitOneClient("chat_msg_ns", cfg.NatsUrl, config.Get().ServerId)
	release := etcdlock.InitDefault(cfg.ETCDClusterList...)
	initialize2.InitLog()
	initialize2.InitRDB()
	// 跨越
	lucky.EnableCrossOrigin()
	// 消息服务
	lucky.Post("/msg/service", func(ctx *context.Context) {
		signature := ctx.FormValue("signature")
		body, err := ctx.GetBody()
		if err != nil {
			log.Error("Read body error %v", err)
			return
		} else {
			// 解密器
			var ec lucky.Encryptor
			if signature == "" {
				if val, ok := encryptors.Load(signature); ok {
					ec = val.(lucky.Encryptor)
				} else {
					if pwd, ep := lucky.ParseLittlePassword(signature); ep == nil {
						ec = lucky.NewLittleCipher(pwd)
						encryptors.Store(signature, ec)
					}
				}
			}
			err = jsp.OnReceivedPackageEC(ctx, body, ec)
			log.Error(err)
		}
	})
	// 文件服务
	lucky.Post("/file/upload", func(ctx *context.Context) {
		//TODO
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
