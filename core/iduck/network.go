package iduck

import (
	"net"
)

type Server interface {
	Run() error
	Handle(conn net.Conn)
}

// 网络同步节点，如游戏房间节点
type INode interface {
	AddConn(interface{}, IConnection)
	DelConn(interface{})
	Serve()
	OnMessage([]byte)
	GetAllMessage() [][][]byte
}

// 网络连接
type IConnection interface {
	ReadMsg()
	WriteMsg(message interface{})
	Close() error
	// 设置自定义数据
	SetData(interface{})
	GetData() interface{}
	// 设置节点
	SetNode(INode)
	GetNode() INode
}
