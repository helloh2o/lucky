# lucky

#### 介绍
这是一个游戏框架，目前支持protobuf消息协议，基于websocket或者socket进行TCP长连接传输，也可以使用可靠的UDP协议KCP进行通信, 支持对消息包加密

数据包加密方式： AES128,AES192,AES256 以及Byte轻量级混淆加密。

数据包读、写、执行逻辑处理分别在各自goroutine中, 可以对单个连接恶意发包进行限制，不会堵塞底层网络。

使用者只需注册消息和消息对应的回调函数，在回调中处理具体逻辑。例如：
```
	Processor.RegisterHandler(code.Hello, &protobuf.Hello{}, logic.Hello)
```

#### 安装教程

go get gitee.com/helloh2o/lucky

#### 使用说明

1. 设置配置参数或保持默认
```
conf.Set(&conf.Data{
		ConnUndoQueueSize:   100,
		ConnWriteQueueSize:  100,
		FirstPackageTimeout: 5,
		ConnReadTimeout:     15,
		ConnWriteTimeout:    5,
		MaxDataPackageSize:  2048,
		MaxHeaderLen:        1024,
	})
```
2. 请参考example下的tcp和websocket 例子

#### TODO
1. 消息JSON 协议
2. mongodb 
3. 帧同步example
#### 欢迎参与

1.  Fork 本仓库
2.  新建 Feat_xxx 分支
3.  提交代码
4.  新建 Pull Request
