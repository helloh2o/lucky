package logic

import (
	"lucky/core/iduck"
	"lucky/example/comm/protobuf"
	"lucky/log"
)

// say hello
func Hello(args ...interface{}) {
	msg := args[0].(*protobuf.Hello)
	log.Debug(msg.Hello)
	conn := args[1].(iduck.IConnection)
	conn.WriteMsg(msg)
}
