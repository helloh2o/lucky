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
var PwdStr = "BH1rStJwNP1YIvNI4Y+8ZVWyqsX47QCTOJTpGLnL2VQHqV0pPu8ZLk3yBc5sRNWmpYjqL2jY9LiFr9EaUsT1Voy3sBadZDKBPQ3g3yP6wOtvrHNxisbuTrPxEHZ6i6sSPAw6mB0rFEsB1OSjXPzlhkmb4lmee1+1aeOgHPaDmUF0vzskwS2iA4TK7ArJ1+fCvWJmY6i2/pDMh1qh3I3PJtBXyBUhET+7w9s5UfcXCVBTQ9beJ1tHC3d5TwgzgkJqkTGkHt1tp2HaTM0fcmd+lY43IP+tsbosJQb7lpqStA94gIlef/AwKnXTQJc1vkZF6Jz5bscCG2CuNhPmKJ8OfA=="

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
