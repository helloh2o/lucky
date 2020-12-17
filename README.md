# Lucky

#### 介绍
一个简洁的游戏/应用网络框架，支持protobuf，JSON 消息协议，基于websocket或者socket(TCP,KCP)进行数据传输, 支持对消息包加密解密。

数据包加密方式： AES128,AES192,AES256 以及Byte轻量级混淆加密。

数据包读、写、执行逻辑处理分别在各自goroutine中, 可以对单个连接恶意发包进行限制超过ConnUndoQueueSize会被忽略，不会堵塞在缓冲区。

使用者只需注册消息和消息对应的回调函数，在回调中处理具体逻辑。例如：
```
        //在处理器上注册（消息码，消息体，消息执行的逻辑代码）
	Processor.RegisterHandler(code.Hello, &protobuf.Hello{}, logic.Hello)
```

#### 安装教程

go get gitee.com/helloh2o/lucky

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
2. 请参考example下的tcp, websocket, kcp, 以及kcp帧同步例子
3. 帧同步部分还需要进一步完善，只是一个基础的实现
4. 聊天室例子, 源码example/chatroom
![Image text](https://gitee.com/helloh2o/lucky/raw/master/example/chatroom/demo.png)

#### TODO
1. mongodb 
#### 欢迎参与

1.  Fork 本仓库
2.  新建 Feat_xxx 分支
3.  提交代码
4.  新建 Pull Request
