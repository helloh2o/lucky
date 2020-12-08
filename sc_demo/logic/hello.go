package logic

import (
	"lucky-day/core/iduck"
	"lucky-day/log"
	"lucky-day/sc_demo/protobuf"
)

// say hello
func Hello(args ...interface{}) {
	msg := args[0].(*protobuf.Hello)
	log.Debug(msg.Hello)
	conn := args[1].(iduck.IConnection)
	conn.WriteMsg(msg)
}
