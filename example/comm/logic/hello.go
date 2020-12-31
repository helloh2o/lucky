package logic

import (
	"github.com/helloh2o/lucky/core/iduck"
	"github.com/helloh2o/lucky/core/iproto"
	"github.com/helloh2o/lucky/example/comm/protobuf"
	"github.com/helloh2o/lucky/log"
)

// Hello say hello logic
func Hello(args ...interface{}) {
	msg := args[iproto.Msg].(*protobuf.Hello)
	log.Debug(msg.Hello)
	conn := args[iproto.Conn].(iduck.IConnection)
	conn.WriteMsg(msg)
}
