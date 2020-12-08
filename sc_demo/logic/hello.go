package logic

import (
	"github.com/sirupsen/logrus"
	"lucky-day/core/duck"
	"lucky-day/sc_demo/protobuf"
)

// say hello
func Hello(args ...interface{}) {
	msg := args[0].(*protobuf.Hello)
	logrus.Println(msg.Hello)
	conn := args[1].(duck.IConnection)
	conn.WriteMsg(msg)
}
