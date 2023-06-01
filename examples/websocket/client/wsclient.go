package main

import (
	"github.com/fasthttp/websocket"
	"github.com/helloh2o/lucky"
	"github.com/helloh2o/lucky/examples/comm/msg"
	"github.com/helloh2o/lucky/examples/comm/msg/code"
	"github.com/helloh2o/lucky/examples/comm/protobuf"
	"github.com/helloh2o/lucky/log"
	"time"
)

func main() {
	/*_, err := log.New("debug", ".", stdlog.LstdFlags|stdlog.Lshortfile)
	if err != nil {
		panic(err)
	}*/
	max := 1000
	for i := 1; i <= max; i++ {
		go runClient(i)
		time.Sleep(time.Millisecond * 100)
	}
	select {}
}

func runClient(id int) {
	hello := protobuf.Hello{Hello: "hello websocket."}
	d := websocket.Dialer{}
	ws, _, err := d.Dial("ws://localhost:2022", nil)
	if err != nil {
		panic(err)
	}
	// 解析协议protobuf
	p := lucky.NewPBProcessor()
	// 内容加密
	msg.SetEncrypt(p)
	i := 1
	p.RegisterHandler(code.Hello, &protobuf.Hello{}, func(args ...interface{}) {
		_msg := args[0].(*protobuf.Hello)
		log.Debug("Id %d, Times %d, msg:: %s", id, i, _msg.Hello)
		i++
		conn := args[1].(lucky.IConnection)
		time.Sleep(time.Millisecond * 200)
		conn.WriteMsg(_msg)
	})
	ic := lucky.NewWSConn(ws, p)
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
