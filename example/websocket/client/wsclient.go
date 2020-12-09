package main

import (
	"github.com/gorilla/websocket"
	"lucky/core/iduck"
	"lucky/core/inet"
	"lucky/core/iproto"
	"lucky/example/comm/msg"
	"lucky/example/comm/msg/code"
	"lucky/example/comm/protobuf"
	"lucky/log"
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
		time.Sleep(time.Millisecond * 100)
	}
	select {}
}

func runClient(id int) {
	hello := protobuf.Hello{Hello: "hello protobuf 3."}
	d := websocket.Dialer{}
	ws, _, err := d.Dial("ws://localhost:2022", nil)
	if err != nil {
		panic(err)
	}
	// 解析协议protobuf
	p := iproto.NewPBProcessor()
	// 内容加密
	msg.SetEncrypt(p)
	i := 1
	p.RegisterHandler(code.Hello, &protobuf.Hello{}, func(args ...interface{}) {
		_msg := args[0].(*protobuf.Hello)
		log.Debug("Id %d, Times %d, msg:: %s", id, i, _msg.Hello)
		i++
		conn := args[1].(iduck.IConnection)
		time.Sleep(time.Millisecond * 200)
		conn.WriteMsg(_msg)
	})
	ic := inet.NewWSConn(ws, p)
	ic.WriteMsg(&hello)
	go func() {
		for {
			_, body, err := ws.ReadMessage()
			if err != nil {
				break
			}
			// throw out the msg
			p.OnReceivedPackage(ic, body)
		}
	}()
}
