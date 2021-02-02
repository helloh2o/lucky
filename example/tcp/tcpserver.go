package main

import (
	"github.com/helloh2o/lucky/core/iduck"
	"github.com/helloh2o/lucky/core/iencrypt/little"
	"github.com/helloh2o/lucky/core/inet"
	"github.com/helloh2o/lucky/core/iproto"
	"github.com/helloh2o/lucky/example/comm/msg/code"
	"github.com/helloh2o/lucky/example/comm/protobuf"
	"github.com/helloh2o/lucky/log"
)

// processor is protobuf processor
var processor = iproto.NewPBProcessor()

func init() {
	//passwd := little.RandPassword()
	passwd := "EyEmxIhoYUFuEc8gDTBlbN/pVOs9Nu/hLCTSjW19Oii0mKNQ9xKPoGJqu1q5Mox4zDT/+MgicJ/j5Nt2sQwK2E8rY3ORVgMUU2v2hmQBb5cP00dettGeF6wvQ36vH2CpGLX9x6RIliP8WAtZqJ0xaT7ixnxxCIr5xRZbutXl8pXqRvSa1+Z/HcuTuFHze4T1ok5A1O4Gge1n6I4ZQjgeHHSSwYs7dQI8oYWQ0MMt3rOywvsVKgUESl2cquDapXrW3PH68MoOPyk1RCe3hxvJNxB3LnLNplVLzkmbTHnZv8AJRedfUoKAJTPsAN0HVzn+XBqUvE2Dvb6nia6tZpmrsA=="
	log.Release("TCP Server password %v", passwd)
	pwd, err := little.ParsePassword(passwd)
	if err != nil {
		panic(err)
	}
	cipher := little.NewCipher(pwd)
	// add encrypt cipher for processor
	processor.SetEncryptor(cipher)
	// 注册消息，以及回调处理
	processor.RegisterHandler(code.Hello, &protobuf.Hello{}, func(args ...interface{}) {
		msg := args[iproto.Msg].(*protobuf.Hello)
		log.Release("Message => from client:: %s", msg.Hello)
		conn := args[iproto.Conn].(iduck.IConnection)
		conn.WriteMsg(msg)
	})
}

func main() {
	// run server
	if s, err := inet.NewTcpServer("localhost:2021", processor); err != nil {
		panic(err)
	} else {
		log.Fatal("%v", s.Run())
	}
}
