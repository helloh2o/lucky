package main

import (
	"github.com/gorilla/websocket"
	"golang.org/x/exp/rand"
	"lucky/cmm/utils"
	"lucky/core/inet"
	"lucky/core/iproto"
	"lucky/example/chatroom/jsonmsg"
	"lucky/log"
	"strconv"
	"time"
)

func main() {
	/*_, err := log.New("debug", ".", stdlog.LstdFlags|stdlog.Lshortfile)
	if err != nil {
		panic(err)
	}*/
	max := 31
	for i := 1; i <= max; i++ {
		go runClient(i)
		time.Sleep(time.Second * 1)
	}
	select {}
}

func runClient(id int) {
	d := websocket.Dialer{}
	ws, _, err := d.Dial("ws://localhost:20220", nil)
	if err != nil {
		panic(err)
	}
	// 解析协议JSON
	p := iproto.NewJSONProcessor()
	wc := inet.NewWSConn(ws, p)
	p.RegisterHandler(jsonmsg.Join_Success, &jsonmsg.JoinRoomSuccess{}, func(args ...interface{}) {
		for {
			_msg := &jsonmsg.ChatMessage{
				FromName: "机器人" + strconv.Itoa(id) + "/" + wc.GetUuid()[:5],
				Content:  utils.RandString(5 + rand.Intn(20)),
			}
			wc.WriteMsg(_msg)
			time.Sleep(time.Second * time.Duration(rand.Intn(10)+1))
			//time.Sleep(time.Millisecond * 200)
		}
	})
	p.RegisterHandler(jsonmsg.Chat_Message, &jsonmsg.ChatMessage{}, func(args ...interface{}) {
		msg := args[iproto.Msg].(*jsonmsg.ChatMessage)
		log.Release("机器人：%s, 发消息：%s", msg.FromName, msg.Content)
	})
	p.RegisterHandler(jsonmsg.Enter_Room, &jsonmsg.EnterRoom{}, nil)
	// 进入房间
	wc.WriteMsg(&jsonmsg.EnterRoom{})
	go func() {
		for {
			_, body, err := ws.ReadMessage()
			if err != nil {
				break
			}
			// throw out the msg
			go p.OnReceivedPackage(wc, body)
		}
	}()
}
