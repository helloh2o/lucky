# Lucky for simple & useful
[**Github**](https://github.com/helloh2o/lucky) <<=>> [**Gitee**](https://gitee.com/helloh2o/lucky)

[![Go Report Card](https://goreportcard.com/badge/github.com/helloh2o/lucky)](https://goreportcard.com/report/github.com/helloh2o/lucky)

#### Introduction

A simple game/application network framework, supports protobuf, JSON message protocol, data transmission based on HTTP/HTTPS, websocket or socket (TCP, KCP, QUIC), supports encryption and decryption of message packets.

Data packet encryption method: AES128, AES192, AES256 and Byte lightweight obfuscated encryption.

Data packet reading, writing, and execution logic processing are respectively in their respective goroutines, which can limit the malicious sending of a single connection and exceed ConnUndoQueueSize, which will be ignored and will not be blocked in the buffer.

The user only needs to register the message and the callback function corresponding to the message, and process the specific logic in the callback. E.g:

```
  //Register on the processor (message code, message body, logic code for message execution)
	Processor.RegisterHandler(code.Hello, &protobuf.Hello{}, logic.Hello)
```

#### Installation tutorial

go get github.com/helloh2o/lucky  or go get gitee.com/helloh2o/lucky

#### how to use it

1.Set configuration parameters or keep default

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

2. Please refer to the http, tcp, websocket, kcp, and kcp frame synchronization examples under the example
3. The frame synchronization part needs to be further improved, it is just a basic realization
4. Chat room example, source code example/chatroom
   ![Image text](https://file.mlog.club/images/2020/12/30/4ee2aca22efb6658502684dfd45a64f1.jpg)

#### Welcome to join :)

1. Welcome to submit PR and issue
2. Open source is not easy, just give a little star if you think it’s good✮
3. This library has been applied to our products

#### Quick Start
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
	passwd := "EyEmxIhoYUFuEc8gDTBlbN/pVOs9Nu/hLCTSjW19Oii0mKNQ9xKPoGJqu1q5Mox4zDT/+MgicJ/j5Nt2sQwK2E8rY3ORVgMUU2v2hmQBb5cP00dettGeF6wvQ36vH2CpGLX9x6RIliP8WAtZqJ0xaT7ixnxxCIr5xRZbutXl8pXqRvSa1+Z/HcuTuFHze4T1ok5A1O4Gge1n6I4ZQjgeHHSSwYs7dQI8oYWQ0MMt3rOywvsVKgUESl2cquDapXrW3PH68MoOPyk1RCe3hxvJNxB3LnLNplVLzkmbTHnZv8AJRedfUoKAJTPsAN0HVzn+XBqUvE2Dvb6nia6tZpmrsA=="
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
	passwd := "EyEmxIhoYUFuEc8gDTBlbN/pVOs9Nu/hLCTSjW19Oii0mKNQ9xKPoGJqu1q5Mox4zDT/+MgicJ/j5Nt2sQwK2E8rY3ORVgMUU2v2hmQBb5cP00dettGeF6wvQ36vH2CpGLX9x6RIliP8WAtZqJ0xaT7ixnxxCIr5xRZbutXl8pXqRvSa1+Z/HcuTuFHze4T1ok5A1O4Gge1n6I4ZQjgeHHSSwYs7dQI8oYWQ0MMt3rOywvsVKgUESl2cquDapXrW3PH68MoOPyk1RCe3hxvJNxB3LnLNplVLzkmbTHnZv8AJRedfUoKAJTPsAN0HVzn+XBqUvE2Dvb6nia6tZpmrsA=="
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
