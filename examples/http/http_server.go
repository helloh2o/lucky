package main

import (
	"github.com/helloh2o/lucky"
	"github.com/helloh2o/lucky/examples/comm/msg"
	"github.com/helloh2o/lucky/examples/comm/msg/code"
	"github.com/helloh2o/lucky/examples/comm/protobuf"
	"github.com/helloh2o/lucky/log"
	"github.com/kataras/iris/v12/context"
)

func main() {
	httpProcessor := lucky.NewPBProcessor()
	msg.SetEncrypt(httpProcessor)
	httpProcessor.RegisterHandler(code.Hello, &protobuf.Hello{}, func(args ...interface{}) {
		hello := args[lucky.Msg].(*protobuf.Hello)
		log.Debug(hello.Hello)
		ctx := args[lucky.Conn].(*context.Context)
		data, err := httpProcessor.WrapMsgNoHeader(hello)
		if err != nil {
			panic(err)
		}
		_, err = ctx.Write(data)
		if err != nil {
			panic(err)
		}
	})
	lucky.EnableCrossOrigin()
	lucky.Post("/", func(context *context.Context) {
		body, err := context.GetBody()
		if err != nil {
			log.Error("Read body error %v", err)
			return
		}
		httpProcessor.OnReceivedPackage(context, body)
	})
	log.Error("http run error %v", lucky.Run(":3001"))
}
