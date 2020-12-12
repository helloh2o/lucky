package logic

import (
	"lucky/core/iduck"
	"lucky/core/iproto"
	"lucky/example/comm/node"
)

// 帧同步开始
func FrameStart(args ...interface{}) {
	conn := args[iproto.Conn].(iduck.IConnection)
	// set test node
	conn.SetNode(node.TestNode)
	node.TestNode.AddConn(conn)
}

// 帧同步开始
func FrameEnd(args ...interface{}) {
	conn := args[iproto.Conn].(iduck.IConnection)
	if data := conn.GetNode(); data != nil {
		_node := data.(iduck.INode)
		_node.Complete()
	}
	_ = conn.Close()
}

func FrameMove(args ...interface{}) {
	conn := args[iproto.Conn].(iduck.IConnection)
	if data := conn.GetNode(); data != nil {
		_node, ok := data.(iduck.INode)
		if ok {
			raw := args[iproto.Raw].([]byte)
			_node.OnRawMessage(raw)
		}
	}
}
