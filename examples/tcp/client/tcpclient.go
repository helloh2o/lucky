package main

import (
	"github.com/helloh2o/lucky"
	"github.com/helloh2o/lucky/examples/comm/msg/code"
	"github.com/helloh2o/lucky/examples/comm/protobuf"
	"github.com/helloh2o/lucky/log"
	"net"
)

// processor is protobuf processor
var processor = lucky.NewPBProcessor()

func init() {
	//passwd := lucky.RandLittlePassword()
	passwd := "EyEmxIhoYUFuEc8gDTBlbN/pVOs9Nu/hLCTSjW19Oii0mKNQ9xKPoGJqu1q5Mox4zDT/+MgicJ/j5Nt2sQwK2E8rY3ORVgMUU2v2hmQBb5cP00dettGeF6wvQ36vH2CpGLX9x6RIliP8WAtZqJ0xaT7ixnxxCIr5xRZbutXl8pXqRvSa1+Z/HcuTuFHze4T1ok5A1O4Gge1n6I4ZQjgeHHSSwYs7dQI8oYWQ0MMt3rOywvsVKgUESl2cquDapXrW3PH68MoOPyk1RCe3hxvJNxB3LnLNplVLzkmbTHnZv8AJRedfUoKAJTPsAN0HVzn+XBqUvE2Dvb6nia6tZpmrsA=="
	log.Release("TCP client password %v", passwd)
	pwd, err := lucky.ParseLittlePassword(passwd)
	if err != nil {
		panic(err)
	}
	cipher := lucky.NewLittleCipher(pwd)
	// add encrypt cipher for processor
	processor.SetEncryptor(cipher)
	// 注册消息，以及回调处理
	processor.RegisterHandler(code.Hello, &protobuf.Hello{}, func(args ...interface{}) {
		msg := args[lucky.Msg].(*protobuf.Hello)
		log.Release("Message => from server:: %s", msg.Hello)
		conn := args[lucky.Conn].(lucky.IConnection)
		_ = conn.Close()
	})
}

func main() {
	conn, err := net.Dial("tcp", "localhost:2021")
	if err != nil {
		panic(err)
	}
	ic := lucky.NewTcpConn(conn, processor)
	ic.WriteMsg(&protobuf.Hello{Hello: "hello lucky."})
	ic.ReadMsg()
}
