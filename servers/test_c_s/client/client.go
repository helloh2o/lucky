package main

import (
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"lucky-day/core/iproto"
	"lucky-day/servers/test_c_s/protobuf_test"
	"net"
)

func main() {
	hello := protobuf_test.Hello{Hello: "hello protobuf 3."}
	hbytes, err := proto.Marshal(&hello)
	if err != nil {
		panic(err)
	}
	protocol := iproto.Protocol{
		Id:      2001,
		Content: hbytes,
	}
	protocolBytes, err := proto.Marshal(&protocol)
	if err != nil {
		panic(err)
	}
	head := make([]byte, 2)
	binary.LittleEndian.PutUint16(head, uint16(len(protocolBytes)))
	pkg := append(head, protocolBytes...)
	conn, err := net.Dial("tcp", "localhost:2021")
	if err != nil {
		panic(err)
	}
	_, err = conn.Write(pkg)
	if err != nil {
		panic(err)
	}
	go func() {
		bf := make([]byte, 2048)
		for {
			// read length
			_, err := io.ReadAtLeast(conn, bf[:2], 2)
			if err != nil {
				logrus.Errorf("TCPConn read message head error %s", err.Error())
				return
			}
			var ln = binary.LittleEndian.Uint16(bf[:2])
			if ln < 1 || ln > 2048 {
				logrus.Errorf("TCPConn message length %d invalid", ln)
				return
			}
			// read data
			_, err = io.ReadFull(conn, bf[:ln])
			if err != nil {
				logrus.Errorf("TCPConn read data err %s", err.Error())
				return
			}
			// throw out the msg
			var p iproto.Protocol
			err = proto.Unmarshal(bf[:ln], &p)
			if err != nil {
				panic(err)
			}
			log.Printf("Client got protocol Id %d", p.Id)
		}
	}()
	select {}
}
