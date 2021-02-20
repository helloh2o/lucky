package logic

import (
	"github.com/helloh2o/lucky"
	"github.com/helloh2o/lucky/examples/comm/node"
)

// FrameStart 帧同步开始
func FrameStart(args ...interface{}) {
	conn := args[lucky.Conn].(lucky.IConnection)
	// set test node
	conn.SetNode(node.TestNode)
	node.TestNode.AddConn(conn)
}

// FrameEnd 帧同步开始
func FrameEnd(args ...interface{}) {
	conn := args[lucky.Conn].(lucky.IConnection)
	if data := conn.GetNode(); data != nil {
		_node := data.(lucky.INode)
		_node.Complete()
	}
	_ = conn.Close()
}

// FrameMove move op
func FrameMove(args ...interface{}) {
	conn := args[lucky.Conn].(lucky.IConnection)
	if data := conn.GetNode(); data != nil {
		_node, ok := data.(lucky.INode)
		if ok {
			raw := args[lucky.Raw].([]byte)
			_node.OnRawMessage(raw)
		}
	}
}
