package main

import (
	"encoding/binary"
	"github.com/helloh2o/lucky/core/iduck"
	"github.com/helloh2o/lucky/core/inet"
	"github.com/helloh2o/lucky/core/iproto"
	"github.com/helloh2o/lucky/example/comm/msg"
	"github.com/helloh2o/lucky/example/comm/msg/code"
	"github.com/helloh2o/lucky/example/comm/protobuf"
	"github.com/helloh2o/lucky/log"
	"github.com/xtaci/kcp-go"
	"io"
	"time"
)

func main() {
	max := 1000
	for i := 1; i <= max; i++ {
		go runClient(i)
		time.Sleep(time.Millisecond * 100)
	}
	select {}
}

func runClient(id int) {
	hello := protobuf.Hello{Hello: "hello kcp 3."}
	conn, err := kcp.DialWithOptions("localhost:2023", nil, 10, 3)
	if err != nil {
		panic(err)
	}
	// 加密
	p := iproto.NewPBProcessor()
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
	ic := inet.NewKcpConn(conn, p)
	ic.WriteMsg(&hello)
	go func() {
		bf := make([]byte, 2048)
		for {
			// read length
			_, err := io.ReadAtLeast(conn, bf[:2], 2)
			if err != nil {
				log.Error("TCPConn read message head error %s", err.Error())
				return
			}
			var ln = binary.LittleEndian.Uint16(bf[:2])
			if ln < 1 || ln > 2048 {
				log.Error("TCPConn message length %d invalid", ln)
				return
			}
			// read data
			_, err = io.ReadFull(conn, bf[:ln])
			if err != nil {
				log.Error("TCPConn read data err %s", err.Error())
				return
			}
			// throw out the msg
			p.OnReceivedPackage(ic, bf[:ln])
		}
	}()
}
