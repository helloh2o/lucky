# Use Lucky to create server by simple, fast, security
[**Github**](https://github.com/helloh2o/lucky) <<=>> [**Gitee**](https://gitee.com/helloh2o/lucky)

[![Go Report Card](https://goreportcard.com/badge/github.com/helloh2o/lucky)](https://goreportcard.com/report/github.com/helloh2o/lucky)

#### [English README](https://github.com/helloh2o/lucky/blob/master/README_en.md)

#### Introduction
A simple and secure game/application network framework, supports protobuf, JSON message protocol, data transmission based on HTTP/HTTPS, websocket or socket (TCP, KCP, QUIC), and supports encryption and decryption of message packets.

Data packet encryption method: AES128, AES192, AES256 and Byte lightweight obfuscated encryption. To help developers and friends write safe and reliable business logic efficiently.

Data packet reading, writing, and execution logic processing are respectively in their respective goroutines. The malicious sending of a single connection can be restricted. If it exceeds ConnUndoQueueSize, it will be ignored and will not stay in the buffer.

The user only needs to register the message and the callback function corresponding to the message, and process the specific logic in the callback. E.g:
```go
    //Register on the processor (message code, message body, logic code for message execution)
Processor.RegisterHandler(code.Hello, &protobuf.Hello{}, logic.Hello)
```

#### Installation tutorial

go get github.com/helloh2o/lucky

#### Instructions for use

1. Set configuration parameters or keep the default
```go
lucky.SetConf(&lucky.Data{
ConnUndoQueueSize: 100,
ConnWriteQueueSize: 100,
FirstPackageTimeout: 5,
ConnReadTimeout: 30,
ConnWriteTimeout: 5,
MaxDataPackageSize: 2048,
MaxHeaderLen: 1024,
})
```
2. Please refer to the http, tcp, websocket, kcp, and kcp frame synchronization examples under the example
3. The frame synchronization part needs to be further improved, which is just a basic realization
4. Chat room example, source code example/chatroom
![Image text](https://raw.githubusercontent.com/helloh2o/lucky/master/examples/chatroom/demo.png)

#### Welcome to participate :)
1. Welcome to submit PR and Issue
2. Open source is not easy, just give a little star if you think it’s good✮
3. The framework has been used in production projects first, DAU≈10w

#### Quick start
> Run as tcp server
```go
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
passwd: = "EyEmxIhoYUFuEc8gDTBlbN / pVOs9Nu / hLCTSjW19Oii0mKNQ9xKPoGJqu1q5Mox4zDT / + MgicJ / j5Nt2sQwK2E8rY3ORVgMUU2v2hmQBb5cP00dettGeF6wvQ36vH2CpGLX9x6RIliP8WAtZqJ0xaT7ixnxxCIr5xRZbutXl8pXqRvSa1 + Z / HcuTuFHze4T1ok5A1O4Gge1n6I4ZQjgeHHSSwYs7dQI8oYWQ0MMt3rOywvsVKgUESl2cquDapXrW3PH68MoOPyk1RCe3hxvJNxB3LnLNplVLzkmbTHnZv8AJRedfUoKAJTPsAN0HVzn + XBqUvE2Dvb6nia6tZpmrsA =="
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
```go
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
passwd: = "EyEmxIhoYUFuEc8gDTBlbN / pVOs9Nu / hLCTSjW19Oii0mKNQ9xKPoGJqu1q5Mox4zDT / + MgicJ / j5Nt2sQwK2E8rY3ORVgMUU2v2hmQBb5cP00dettGeF6wvQ36vH2CpGLX9x6RIliP8WAtZqJ0xaT7ixnxxCIr5xRZbutXl8pXqRvSa1 + Z / HcuTuFHze4T1ok5A1O4Gge1n6I4ZQjgeHHSSwYs7dQI8oYWQ0MMt3rOywvsVKgUESl2cquDapXrW3PH68MoOPyk1RCe3hxvJNxB3LnLNplVLzkmbTHnZv8AJRedfUoKAJTPsAN0HVzn + XBqUvE2Dvb6nia6tZpmrsA =="
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
