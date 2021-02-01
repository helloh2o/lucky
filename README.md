# Lucky for simple & useful
[**Github**](https://github.com/helloh2o/lucky) <<=>> [**Gitee**](https://gitee.com/helloh2o/lucky)

[![Go Report Card](https://goreportcard.com/badge/github.com/helloh2o/lucky)](https://goreportcard.com/report/github.com/helloh2o/lucky)

#### [English README](https://github.com/helloh2o/lucky/blob/master/README_en.md)

#### 介绍
一个简洁的游戏/应用网络框架，支持protobuf，JSON 消息协议，基于HTTP/HTTPS,websocket或者socket(TCP,KCP,QUIC)进行数据传输, 支持对消息包加密解密。

数据包加密方式： AES128,AES192,AES256 以及Byte轻量级混淆加密。

数据包读、写、执行逻辑处理分别在各自goroutine中, 可以对单个连接恶意发包进行限制超过ConnUndoQueueSize会被忽略，不停留在缓冲区。

使用者只需注册消息和消息对应的回调函数，在回调中处理具体逻辑。例如：
```
    //在处理器上注册（消息码，消息体，消息执行的逻辑代码）
	Processor.RegisterHandler(code.Hello, &protobuf.Hello{}, logic.Hello)
```

#### 安装教程

go get github.com/helloh2o/lucky  或者 go get gitee.com/helloh2o/lucky

#### 使用说明

1. 设置配置参数或保持默认
```
conf.Set(&conf.Data{
		ConnUndoQueueSize:   100,
		FirstPackageTimeout: 5,
		ConnReadTimeout:     35,
		ConnWriteTimeout:    5,
		MaxDataPackageSize:  4096,
		MaxHeaderLen:        1024,
	})
```
2. 请参考example下的http, tcp, websocket, kcp, 以及kcp帧同步例子
3. 帧同步部分还需要进一步完善，只是一个基础的实现
4. 聊天室例子, 源码example/chatroom
![Image text](https://gitee.com/helloh2o/lucky/raw/master/example/chatroom/demo.png)

#### 欢迎参与 :)
1. 欢迎提交PR 和 Issue
2. 开源不易，觉得不错就给个小星星✮吧 
3. 该框架已先在生产项目中使用，DAU≈10w

#### 快速开始
> TCP Server
```
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
	// register msg and it's callback
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
```

> TCP Client
```
package main

import (
	"github.com/helloh2o/lucky/core/iduck"
	"github.com/helloh2o/lucky/core/iencrypt/little"
	"github.com/helloh2o/lucky/core/inet"
	"github.com/helloh2o/lucky/core/iproto"
	"github.com/helloh2o/lucky/example/comm/msg/code"
	"github.com/helloh2o/lucky/example/comm/protobuf"
	"github.com/helloh2o/lucky/log"
	"net"
)

// processor is protobuf processor
var processor = iproto.NewPBProcessor()

func init() {
	//passwd := little.RandPassword()
	passwd := "BH1rStJwNP1YIvNI4Y+8ZVWyqsX47QCTOJTpGLnL2VQHqV0pPu8ZLk3yBc5sRNWmpYjqL2jY9LiFr9EaUsT1Voy3sBadZDKBPQ3g3yP6wOtvrHNxisbuTrPxEHZ6i6sSPAw6mB0rFEsB1OSjXPzlhkmb4lmee1+1aeOgHPaDmUF0vzskwS2iA4TK7ArJ1+fCvWJmY6i2/pDMh1qh3I3PJtBXyBUhET+7w9s5UfcXCVBTQ9beJ1tHC3d5TwgzgkJqkTGkHt1tp2HaTM0fcmd+lY43IP+tsbosJQb7lpqStA94gIlef/AwKnXTQJc1vkZF6Jz5bscCG2CuNhPmKJ8OfA=="
	log.Release("TCP client password %v", passwd)
	pwd, err := little.ParsePassword(passwd)
	if err != nil {
		panic(err)
	}
	cipher := little.NewCipher(pwd)
	// add encrypt cipher for processor
	processor.SetEncryptor(cipher)
	// register msg and it's callback
	processor.RegisterHandler(code.Hello, &protobuf.Hello{}, func(args ...interface{}) {
		msg := args[iproto.Msg].(*protobuf.Hello)
		log.Release("Message => from server:: %s", msg.Hello)
		conn := args[iproto.Conn].(iduck.IConnection)
		_ = conn.Close()
	})
}

func main() {
	conn, err := net.Dial("tcp", "localhost:2021")
	if err != nil {
		panic(err)
	}
	ic := inet.NewTcpConn(conn, processor)
	ic.WriteMsg(&protobuf.Hello{Hello: "hello lucky."})
	ic.ReadMsg()
}

```
