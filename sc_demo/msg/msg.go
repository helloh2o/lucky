package msg

import (
	"lucky-day/core/iproto"
	"lucky-day/sc_demo/logic"
	"lucky-day/sc_demo/msg/code"
	"lucky-day/sc_demo/protobuf"
)

var Processor = iproto.NewPBProcessor()

// 注册逻辑
func init() {
	Processor.RegisterHandler(code.Hello, &protobuf.Hello{}, logic.Hello)
}
