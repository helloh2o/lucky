package main

import (
	"encoding/binary"
	"io"
	stdlog "log"
	"lucky-day/core/iduck"
	"lucky-day/core/iencrypt/little"
	"lucky-day/core/inet"
	"lucky-day/core/iproto"
	"lucky-day/example/tcp_encrypt/msg/code"
	"lucky-day/example/tcp_encrypt/protobuf"
	"lucky-day/log"
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
	clientSetEncrypt(p)
	i := 1
	p.RegisterHandler(code.Hello, &protobuf.Hello{}, func(args ...interface{}) {
		_msg := args[0].(*protobuf.Hello)
		log.Debug("Id %d, Times %d, msg:: %s", id, i, _msg.Hello)
		i++
		conn := args[1].(iduck.IConnection)
		time.Sleep(time.Millisecond * 200)
		conn.WriteMsg(_msg)
	})
	ic := inet.NewTcpConn(conn, p, 100)
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

func clientSetEncrypt(p iduck.Processor) {
	//pwdStr := little.RandPassword()
	pwdStr := "BH1rStJwNP1YIvNI4Y+8ZVWyqsX47QCTOJTpGLnL2VQHqV0pPu8ZLk3yBc5sRNWmpYjqL2jY9LiFr9EaUsT1Voy3sBadZDKBPQ3g3yP6wOtvrHNxisbuTrPxEHZ6i6sSPAw6mB0rFEsB1OSjXPzlhkmb4lmee1+1aeOgHPaDmUF0vzskwS2iA4TK7ArJ1+fCvWJmY6i2/pDMh1qh3I3PJtBXyBUhET+7w9s5UfcXCVBTQ9beJ1tHC3d5TwgzgkJqkTGkHt1tp2HaTM0fcmd+lY43IP+tsbosJQb7lpqStA94gIlef/AwKnXTQJc1vkZF6Jz5bscCG2CuNhPmKJ8OfA=="
	pwd, err := little.ParsePassword(pwdStr)
	if err != nil {
		panic(err)
	}
	p.SetEncrypt(little.NewCipher(pwd))
}
