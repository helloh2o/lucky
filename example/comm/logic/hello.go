package logic

import (
	"lucky/core/iduck"
	"lucky/core/iproto"
	"lucky/example/comm/protobuf"
	"lucky/log"
)

// say hello
func Hello(args ...interface{}) {
	msg := args[iproto.Msg].(*protobuf.Hello)
	log.Debug(msg.Hello)
	conn := args[iproto.Conn].(iduck.IConnection)
	conn.WriteMsg(msg)
}
