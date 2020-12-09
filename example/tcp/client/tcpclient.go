package main

import (
	"encoding/binary"
	"io"
	stdlog "log"
	"lucky/core/iduck"
	"lucky/core/inet"
	"lucky/core/iproto"
	"lucky/example/comm/msg"
	"lucky/example/comm/msg/code"
	"lucky/example/comm/protobuf"
	"lucky/log"
	"net"
	"time"
)

func main() {
	_, err := log.New("debug", ".", stdlog.LstdFlags|stdlog.Lshortfile)
	if err != nil {
		panic(err)
	}
	max := 1000
	for i := 1; i <= max; i++ {
		go runClient(i)
		time.Sleep(time.Millisecond * 100)
	}
	select {}
}

func runClient(id int) {
	hello := protobuf.Hello{Hello: "hello protobuf 3."}
	/*hbytes, err := proto.Marshal(&hello)
	if err != nil {
		panic(err)
	}
	protocol := iproto.Protocol{
		Id:      2001,
		Content: hbytes,
	}*/
	/*protocolBytes, err := proto.Marshal(&protocol)
	if err != nil {
		panic(err)
	}
	head := make([]byte, 2)
	binary.LittleEndian.PutUint16(head, uint16(len(protocolBytes)))
	pkg := append(head, protocolBytes...)*/
	conn, err := net.Dial("tcp", "localhost:2021")
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
	ic := inet.NewTcpConn(conn, p)
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
