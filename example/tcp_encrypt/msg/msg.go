package msg

import (
	"lucky-day/core/iduck"
	"lucky-day/core/iencrypt/little"
	"lucky-day/core/iproto"
	"lucky-day/example/tcp_encrypt/logic"
	"lucky-day/example/tcp_encrypt/msg/code"
	"lucky-day/example/tcp_encrypt/protobuf"
)

var Processor = iproto.NewPBProcessor()

func SetEncrypt(p iduck.Processor) {
	//pwdStr := little.RandPassword()
	pwdStr := "BH1rStJwNP1YIvNI4Y+8ZVWyqsX47QCTOJTpGLnL2VQHqV0pPu8ZLk3yBc5sRNWmpYjqL2jY9LiFr9EaUsT1Voy3sBadZDKBPQ3g3yP6wOtvrHNxisbuTrPxEHZ6i6sSPAw6mB0rFEsB1OSjXPzlhkmb4lmee1+1aeOgHPaDmUF0vzskwS2iA4TK7ArJ1+fCvWJmY6i2/pDMh1qh3I3PJtBXyBUhET+7w9s5UfcXCVBTQ9beJ1tHC3d5TwgzgkJqkTGkHt1tp2HaTM0fcmd+lY43IP+tsbosJQb7lpqStA94gIlef/AwKnXTQJc1vkZF6Jz5bscCG2CuNhPmKJ8OfA=="
	pwd, err := little.ParsePassword(pwdStr)
	if err != nil {
		panic(err)
	}
	p.SetEncrypt(little.NewCipher(pwd))
}
func init() {
	// 设置加密器
	SetEncrypt(Processor)
	// 注册逻辑
	Processor.RegisterHandler(code.Hello, &protobuf.Hello{}, logic.Hello)
}
