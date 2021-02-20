package logic

import (
	"github.com/helloh2o/lucky"
	"github.com/helloh2o/lucky/examples/comm/protobuf"
	"github.com/helloh2o/lucky/log"
)

// Hello say hello logic
func Hello(args ...interface{}) {
	msg := args[lucky.Msg].(*protobuf.Hello)
	log.Debug(msg.Hello)
	conn := args[lucky.Conn].(lucky.IConnection)
	conn.WriteMsg(msg)
}
