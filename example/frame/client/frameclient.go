package main

import (
	"encoding/binary"
	"github.com/helloh2o/lucky/conf"
	"github.com/helloh2o/lucky/core/iduck"
	"github.com/helloh2o/lucky/core/inet"
	"github.com/helloh2o/lucky/core/iproto"
	"github.com/helloh2o/lucky/example/comm/msg"
	"github.com/helloh2o/lucky/example/comm/msg/code"
	"github.com/helloh2o/lucky/example/comm/protobuf"
	"github.com/helloh2o/lucky/log"
	"github.com/xtaci/kcp-go"
	"io"
	"math/rand"
	"sync"
	"time"
)

var wg sync.WaitGroup

func main() {
	max := 100
	wg.Add(max)
	for i := 1; i <= max; i++ {
		go runClient(i)
		time.Sleep(time.Millisecond * 100)
	}
	wg.Wait()
	log.Release("======== frame battle complete ========")
}

func runClient(id int) {
	hello := protobuf.Hello{Hello: "hello kcp frame."}
	conn, err := kcp.DialWithOptions("localhost:2024", nil, 10, 3)
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
		conn.WriteMsg(&protobuf.CsStartFrame{})
		// 1分钟后结束同步, 移动操作
		go func() {
			defer wg.Done()
			time.Sleep(time.Second * 5)
			for i := 0; i < 60; i++ {
				conn.WriteMsg(&protobuf.CsMove{
					FromX: 1,
					FromY: 2.1,
					FromZ: 3,
					ToX:   3,
					ToY:   9,
					ToZ:   41.2,
					Speed: 9,
				})
				sn := rand.Intn(800) + 100
				time.Sleep(time.Millisecond * time.Duration(sn))
			}
			conn.WriteMsg(&protobuf.CsEndFrame{})
		}()
	})
	p.RegisterHandler(code.FrameStart, &protobuf.CsStartFrame{}, nil)
	p.RegisterHandler(code.FrameEnd, &protobuf.CsEndFrame{}, nil)
	p.RegisterHandler(code.MoveOp, &protobuf.CsMove{}, nil)
	p.RegisterHandler(code.FrameData, &iproto.ScFrame{}, func(args ...interface{}) {
		f := args[0].(*iproto.ScFrame)
		log.Release("==== Received FrameData message packages length %d ====", len(f.Protocols))
	})
	ic := inet.NewKcpConn(conn, p)
	ic.WriteMsg(&hello)
	go func() {
		bf := make([]byte, 8192)
		for {
			// read length
			_, err := io.ReadAtLeast(conn, bf[:2], 2)
			if err != nil {
				log.Error("TCPConn read message head error %s", err.Error())
				return
			}
			var ln = binary.LittleEndian.Uint16(bf[:2])
			if ln < 1 || ln > uint16(conf.C.MaxDataPackageSize) {
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
