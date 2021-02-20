package lucky

import (
	"net"
)

// Server interface
type Server interface {
	Run() error
	Handle(conn net.Conn)
}

// INode 网络同步节点，如游戏房间节点,聊天室节点
type INode interface {
	AddConn(IConnection) error
	DelConn(string) error
	Serve()
	OnRawMessage([]byte) error
	OnProtocolMessage(interface{}) error
	GetAllMessage() chan []interface{}
	Destroy() error
	Complete() error
}

// IConnection 网络连接
type IConnection interface {
	GetUuid() string
	ReadMsg()
	WriteMsg(message interface{})
	Close() error
	AfterClose(func())
	// 设置自定义数据
	SetData(interface{})
	GetData() interface{}
	// 设置节点
	SetNode(INode)
	GetNode() INode
	// 是否关闭
	IsClosed() bool
}
