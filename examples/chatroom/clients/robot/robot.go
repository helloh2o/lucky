package main

import (
	"github.com/fasthttp/websocket"
	"github.com/helloh2o/lucky"
	"github.com/helloh2o/lucky/examples/chatroom/jsonmsg"
	"github.com/helloh2o/lucky/log"
	"github.com/helloh2o/lucky/utils"
	"golang.org/x/exp/rand"
	"strconv"
	"sync/atomic"
	"time"
)

func main() {
	/*_, err := log.New("debug", ".", stdlog.LstdFlags|stdlog.Lshortfile)
	if err != nil {
		panic(err)
	}*/
	max := 100
	for i := 1; i <= max; i++ {
		go runClient(i)
		time.Sleep(time.Millisecond * 300)
	}
	select {}
}

func runClient(id int) {
	var chatting int64
	d := websocket.Dialer{}
	ws, _, err := d.Dial("ws://localhost:20220", nil)
	if err != nil {
		panic(err)
	}
	// 解析协议JSON
	p := lucky.NewJSONProcessor()
	wc := lucky.NewWSConn(ws, p)
	p.RegisterHandler(jsonmsg.JoinSuccessCode, &jsonmsg.JoinRoomSuccess{}, func(args ...interface{}) {
		go func() {
			atomic.AddInt64(&chatting, 1)
			for {
				_msg := &jsonmsg.ChatMessage{
					FromName: "机器人" + strconv.Itoa(id) + "/" + wc.GetUuid()[:5],
					Content:  utils.RandString(5 + rand.Intn(20)),
				}
				wc.WriteMsg(_msg)
				time.Sleep(time.Second * time.Duration(rand.Intn(10)+1))
			}
		}()
	})
	p.RegisterHandler(jsonmsg.ChatMessageCode, &jsonmsg.ChatMessage{}, func(args ...interface{}) {
		msg := args[lucky.Msg].(*jsonmsg.ChatMessage)
		log.Release("机器人：%s, 发消息：%s", msg.FromName, msg.Content)
	})
	p.RegisterHandler(jsonmsg.EnterRoomCode, &jsonmsg.EnterRoom{}, nil)
	// 进入房间
	wc.WriteMsg(&jsonmsg.EnterRoom{})
	go func() {
		for {
			_, body, err := ws.ReadMessage()
			if err != nil {
				break
			}
			if atomic.LoadInt64(&chatting) == 0 {
				// throw out the msg
				p.OnReceivedPackage(wc, body)
			} else {
				// do not handle message on chatting, but read TCP cache buff
			}
		}
	}()
}
