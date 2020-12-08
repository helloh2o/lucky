package logic

import (
	"lucky-day/core/iduck"
	"lucky-day/example/tcp_encrypt/protobuf"
	"lucky-day/log"
)

// say hello
func Hello(args ...interface{}) {
	msg := args[0].(*protobuf.Hello)
	log.Debug(msg.Hello)
	conn := args[1].(iduck.IConnection)
	conn.WriteMsg(msg)
}
