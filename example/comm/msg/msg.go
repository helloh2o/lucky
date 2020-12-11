package msg

import (
	"lucky/core/iduck"
	"lucky/core/iencrypt/aes"
	"lucky/core/iencrypt/little"
	"lucky/core/iproto"
	"lucky/example/comm/logic"
	"lucky/example/comm/msg/code"
	"lucky/example/comm/protobuf"
)

var Processor = iproto.NewPBProcessor()

func SetEncrypt(p iduck.Processor) {
	//pwdStr := little.RandPassword()
	pwdStr := "BH1rStJwNP1YIvNI4Y+8ZVWyqsX47QCTOJTpGLnL2VQHqV0pPu8ZLk3yBc5sRNWmpYjqL2jY9LiFr9EaUsT1Voy3sBadZDKBPQ3g3yP6wOtvrHNxisbuTrPxEHZ6i6sSPAw6mB0rFEsB1OSjXPzlhkmb4lmee1+1aeOgHPaDmUF0vzskwS2iA4TK7ArJ1+fCvWJmY6i2/pDMh1qh3I3PJtBXyBUhET+7w9s5UfcXCVBTQ9beJ1tHC3d5TwgzgkJqkTGkHt1tp2HaTM0fcmd+lY43IP+tsbosJQb7lpqStA94gIlef/AwKnXTQJc1vkZF6Jz5bscCG2CuNhPmKJ8OfA=="
	pwd, err := little.ParsePassword(pwdStr)
	if err != nil {
		panic(err)
	}
	// 混淆加密
	//cipher := little.NewCipher(pwd)
	// 高级标准加密
	cipher := aes.NewAESCipher(pwdStr)
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
