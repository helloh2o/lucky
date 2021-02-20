package msg

import (
	"github.com/helloh2o/lucky"
	"github.com/helloh2o/lucky/examples/comm/logic"
	"github.com/helloh2o/lucky/examples/comm/msg/code"
	"github.com/helloh2o/lucky/examples/comm/protobuf"
)

// Processor is message handler
var Processor = lucky.NewPBProcessor()

// PwdStr is encrypt key
var PwdStr = "EyEmxIhoYUFuEc8gDTBlbN/pVOs9Nu/hLCTSjW19Oii0mKNQ9xKPoGJqu1q5Mox4zDT/+MgicJ/j5Nt2sQwK2E8rY3ORVgMUU2v2hmQBb5cP00dettGeF6wvQ36vH2CpGLX9x6RIliP8WAtZqJ0xaT7ixnxxCIr5xRZbutXl8pXqRvSa1+Z/HcuTuFHze4T1ok5A1O4Gge1n6I4ZQjgeHHSSwYs7dQI8oYWQ0MMt3rOywvsVKgUESl2cquDapXrW3PH68MoOPyk1RCe3hxvJNxB3LnLNplVLzkmbTHnZv8AJRedfUoKAJTPsAN0HVzn+XBqUvE2Dvb6nia6tZpmrsA=="

// SetEncrypt for processor
func SetEncrypt(p lucky.Processor) {
	//pwdStr := little.RandLittlePassword()
	pwd, err := lucky.ParseLittlePassword(PwdStr)
	if err != nil {
		panic(err)
	}
	// 混淆加密
	cipher := lucky.NewLittleCipher(pwd)
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
	Processor.RegisterHandler(code.FrameData, &lucky.ScFrame{}, nil)
	Processor.RegisterHandler(code.FrameEnd, &protobuf.CsEndFrame{}, logic.FrameEnd)
	Processor.RegisterHandler(code.MoveOp, &protobuf.CsMove{}, logic.FrameMove)
}
