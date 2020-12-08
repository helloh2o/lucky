package iduck

import (
	"net"
)

type Server interface {
	Run() error
	Handle(conn net.Conn)
}

type IConnection interface {
	ReadMsg()
	WriteMsg(message interface{})
	Close() error
}
