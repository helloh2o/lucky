package main

import (
	"flag"
	"github.com/helloh2o/lucky"
	"github.com/helloh2o/lucky/examples/chatroom/jsonmsg"
	"github.com/helloh2o/lucky/im/server/config"
	"github.com/helloh2o/lucky/im/server/constants"
	"github.com/helloh2o/lucky/im/server/immsg"
	"github.com/helloh2o/lucky/im/server/route"
	"github.com/helloh2o/lucky/log"
	"github.com/helloh2o/lucky/natsq"
	"github.com/helloh2o/lucky/utils"
	"github.com/helloh2o/lucky/utils/etcdlock"
	"github.com/nats-io/stan.go"
)

var confPath = flag.String("conf", "./config.yaml", "config file path")

func main() {
	flag.Parse()
	cfg := config.Initialize(*confPath)
	jsp := route.JsonHandler()
	_ = natsq.InitOneClient("chat_msg_ns", cfg.NatsUrl, config.Get().ServerId)
	release := etcdlock.InitDefault(cfg.ETCDClusterList...)
	jsp.RegisterHandler(constants.GATE_MSG, &immsg.ConnectMsg{}, func(args ...interface{}) {
		// 链接建立，订阅自己的频道
		msg := args[lucky.Msg].(*immsg.ConnectMsg)
		conn := args[lucky.Conn].(lucky.IConnection)
		var subscriptionSelf stan.Subscription
		if msg.Type == constants.Connecting {
			// 订阅自己的频道消息
			subscriptionSelf = natsq.SubscribeDurable(msg.UserId, msg.UserId, func(m *stan.Msg) {
				// 转发消息
				if err := conn.WriteMsg(m); err != nil {
					log.Error(err.Error())
					// 关闭订阅
					_ = subscriptionSelf.Close()
				}
			})
			for _, gId := range msg.Groups {
				var subGroup stan.Subscription
				// 订阅群组消息
				subGroup = natsq.SubscribeDurable(gId, gId, func(m *stan.Msg) {
					// 转发消息
					if err := conn.WriteMsg(m); err != nil {
						log.Error(err.Error())
						// 关闭群组订阅
						_ = subGroup.Close()
					}
				})
			}
			// 回去消息
			_ = conn.WriteMsg(&immsg.BaseMsg{Type: constants.Connected})
		}
	})
	lucky.SetConf(&lucky.Data{
		ConnUndoQueueSize:   100,
		FirstPackageTimeout: 5,
		ConnReadTimeout:     15,
		ConnWriteTimeout:    10,
		MaxDataPackageSize:  2048,
		MaxHeaderLen:        1024,
	})
	go func() {
		if s, err := lucky.NewWsServer(config.Get().ListenAddr, jsonmsg.Processor); err != nil {
			panic(err)
		} else {
			log.Fatal(s.Run())
		}
	}()
	// 优雅退出 made test on ctrl + C  | killall proc | service restart
	graceDone := make(chan struct{})
	go utils.GraceExit(graceDone, release)
	<-graceDone
}
