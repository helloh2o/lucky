package msg

import (
	"github.com/helloh2o/lucky/core/iduck"
	"github.com/helloh2o/lucky/core/iencrypt/little"
	"github.com/helloh2o/lucky/core/iproto"
	"github.com/helloh2o/lucky/example/comm/logic"
	"github.com/helloh2o/lucky/example/comm/msg/code"
	"github.com/helloh2o/lucky/example/comm/protobuf"
)

// Processor is message handler
var Processor = iproto.NewPBProcessor()

// PwdStr is encrypt key
var PwdStr = "EyEmxIhoYUFuEc8gDTBlbN/pVOs9Nu/hLCTSjW19Oii0mKNQ9xKPoGJqu1q5Mox4zDT/+MgicJ/j5Nt2sQwK2E8rY3ORVgMUU2v2hmQBb5cP00dettGeF6wvQ36vH2CpGLX9x6RIliP8WAtZqJ0xaT7ixnxxCIr5xRZbutXl8pXqRvSa1+Z/HcuTuFHze4T1ok5A1O4Gge1n6I4ZQjgeHHSSwYs7dQI8oYWQ0MMt3rOywvsVKgUESl2cquDapXrW3PH68MoOPyk1RCe3hxvJNxB3LnLNplVLzkmbTHnZv8AJRedfUoKAJTPsAN0HVzn+XBqUvE2Dvb6nia6tZpmrsA=="

// SetEncrypt for processor
func SetEncrypt(p iduck.Processor) {
	//pwdStr := little.RandPassword()
	pwd, err := little.ParsePassword(PwdStr)
	if err != nil {
		panic(err)
	}
	// 混淆加密
	cipher := little.NewCipher(pwd)
	//cipher := oor.NewXORCipher("BH1rStJwNP1Y%d^*IvNI4Y+8ZVWyqsX")
	// 高级标准加密
	//cipher := aes.NewAESCipher(pwdStr)
	_ = pwd
	p.SetEncryptor(cipher)
}
func init() {
	// 注册消息，以及回调处理
	Processor.RegisterHandler(code.Hello, &protobuf.Hello{}, logic.Hello)

	// 帧同步处理
	Processor.RegisterHandler(code.FrameStart, &protobuf.CsStartFrame{}, logic.FrameStart)
	Processor.RegisterHandler(code.FrameData, &iproto.ScFrame{}, nil)
	Processor.RegisterHandler(code.FrameEnd, &protobuf.CsEndFrame{}, logic.FrameEnd)
	Processor.RegisterHandler(code.MoveOp, &protobuf.CsMove{}, logic.FrameMove)
}
