# Use Lucky to create server by simple, fast, security
[**Github**](https://github.com/helloh2o/lucky) <<=>> [**Gitee**](https://gitee.com/helloh2o/lucky)

[![Go Report Card](https://goreportcard.com/badge/github.com/helloh2o/lucky)](https://goreportcard.com/report/github.com/helloh2o/lucky)

#### [English README](https://github.com/helloh2o/lucky/blob/master/README_en.md)

#### 介绍
一个简洁安全游戏/应用网络框架，支持protobuf，JSON 消息协议，基于HTTP/HTTPS,websocket或者socket(TCP,KCP,QUIC)进行数据传输, 支持对消息包加密解密。

数据包加密方式： AES128,AES192,AES256 以及Byte轻量级混淆加密。以帮助开发者朋友能高效率写出安全可靠的业务逻辑。

数据包读、写、执行逻辑处理分别在各自goroutine中, 可以对单个连接恶意发包进行限制超过ConnUndoQueueSize会被忽略，不停留在缓冲区。

使用者只需注册消息和消息对应的回调函数，在回调中处理具体逻辑。例如：
```
    //在处理器上注册（消息码，消息体，消息执行的逻辑代码）
	Processor.RegisterHandler(code.Hello, &protobuf.Hello{}, logic.Hello)
```

#### 安装教程

go get github.com/helloh2o/lucky

#### 使用说明

1. 设置配置参数或保持默认
```
lucky.SetConf(&lucky.Data{
		ConnUndoQueueSize:   100,
		ConnWriteQueueSize:  100,
		FirstPackageTimeout: 5,
		ConnReadTimeout:     30,
		ConnWriteTimeout:    5,
		MaxDataPackageSize:  2048,
		MaxHeaderLen:        1024,
	})
```
2. 请参考example下的http, tcp, websocket, kcp, 以及kcp帧同步例子
3. 帧同步部分还需要进一步完善，只是一个基础的实现
4. 聊天室例子, 源码example/chatroom
![Image text](https://raw.githubusercontent.com/helloh2o/lucky/master/examples/chatroom/demo.png)

#### 欢迎参与 :)
1. 欢迎提交PR 和 Issue
2. 开源不易，觉得不错就给个小星星✮吧 
3. 该框架已先在生产项目中使用，DAU≈10w

#### 快速开始
> Run as tcp server
```
package main

import (
	"github.com/helloh2o/lucky"
	"github.com/helloh2o/lucky/examples/comm/msg/code"
	"github.com/helloh2o/lucky/examples/comm/protobuf"
	"github.com/helloh2o/lucky/log"
)

// processor is protobuf processor
var processor = lucky.NewPBProcessor()

func init() {
	//passwd := little.RandLittlePassword()
	passwd := "EyEmxIhoYUFuEc8gDTBlbN/pVOs9Nu/hLCTSjW19Oii0mKNQ9xKPoGJqu1q5Mox4zDT/+MgicJ/j5Nt2sQwK2E8rY3ORVgMUU2v2hmQBb5cP00dettGeF6wvQ36vH2CpGLX9x6RIliP8WAtZqJ0xaT7ixnxxCIr5xRZbutXl8pXqRvSa1+Z/HcuTuFHze4T1ok5A1O4Gge1n6I4ZQjgeHHSSwYs7dQI8oYWQ0MMt3rOywvsVKgUESl2cquDapXrW3PH68MoOPyk1RCe3hxvJNxB3LnLNplVLzkmbTHnZv8AJRedfUoKAJTPsAN0HVzn+XBqUvE2Dvb6nia6tZpmrsA=="
	log.Release("TCP Server password %v", passwd)
	pwd, err := lucky.ParseLittlePassword(passwd)
	if err != nil {
		panic(err)
	}
	cipher := lucky.NewLittleCipher(pwd)
	// add encrypt cipher for processor
	processor.SetEncryptor(cipher)
	// register message && callback
	processor.RegisterHandler(code.Hello, &protobuf.Hello{}, func(args ...interface{}) {
		msg := args[lucky.Msg].(*protobuf.Hello)
		log.Release("Message => from client:: %s", msg.Hello)
		conn := args[lucky.Conn].(lucky.IConnection)
		conn.WriteMsg(msg)
	})
}

func main() {
	// run server
	if s, err := lucky.NewTcpServer("localhost:2021", processor); err != nil {
		panic(err)
	} else {
		log.Fatal("%v", s.Run())
	}
}

```

> go tcp client
```
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
	// register message && callback
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

```
