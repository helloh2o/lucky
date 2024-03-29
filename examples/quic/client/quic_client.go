package main

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"github.com/helloh2o/lucky"
	"github.com/helloh2o/lucky/examples/comm/msg"
	"github.com/helloh2o/lucky/examples/comm/msg/code"
	"github.com/helloh2o/lucky/examples/comm/protobuf"
	"github.com/helloh2o/lucky/log"
	"github.com/quic-go/quic-go"
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

func runClient(id int) error {
	hello := protobuf.Hello{Hello: "hello kcp 3."}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-hello-example"},
	}
	session, err := quic.DialAddr(context.Background(), "localhost:2024", tlsConfig, nil)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	stream, err := session.OpenStreamSync(context.Background())
	if err != nil {
		log.Error(err.Error())
		return err
	}

	// 加密
	p := lucky.NewPBProcessor()
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
	ic := lucky.NewQuicStream(stream, p)
	ic.WriteMsg(&hello)
	go func() {
		bf := make([]byte, 2048)
		for {
			// read length
			_, err := io.ReadAtLeast(stream, bf[:2], 2)
			if err != nil {
				log.Error("quic stream read message head error %s", err.Error())
				return
			}
			var ln = binary.LittleEndian.Uint16(bf[:2])
			if ln < 1 || ln > 2048 {
				log.Error("quic stream message length %d invalid", ln)
				return
			}
			// read data
			_, err = io.ReadFull(stream, bf[:ln])
			if err != nil {
				log.Error("quic stream read data err %s", err.Error())
				return
			}
			// throw out the msg
			p.OnReceivedPackage(ic, bf[:ln])
		}
	}()
	return nil
}
