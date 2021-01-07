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
	passwd := "BH1rStJwNP1YIvNI4Y+8ZVWyqsX47QCTOJTpGLnL2VQHqV0pPu8ZLk3yBc5sRNWmpYjqL2jY9LiFr9EaUsT1Voy3sBadZDKBPQ3g3yP6wOtvrHNxisbuTrPxEHZ6i6sSPAw6mB0rFEsB1OSjXPzlhkmb4lmee1+1aeOgHPaDmUF0vzskwS2iA4TK7ArJ1+fCvWJmY6i2/pDMh1qh3I3PJtBXyBUhET+7w9s5UfcXCVBTQ9beJ1tHC3d5TwgzgkJqkTGkHt1tp2HaTM0fcmd+lY43IP+tsbosJQb7lpqStA94gIlef/AwKnXTQJc1vkZF6Jz5bscCG2CuNhPmKJ8OfA=="
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
