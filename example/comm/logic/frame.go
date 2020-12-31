package logic

import (
	"github.com/helloh2o/lucky/core/iduck"
	"github.com/helloh2o/lucky/core/iproto"
	"github.com/helloh2o/lucky/example/comm/node"
)

// FrameStart 帧同步开始
func FrameStart(args ...interface{}) {
	conn := args[iproto.Conn].(iduck.IConnection)
	// set test node
	conn.SetNode(node.TestNode)
	node.TestNode.AddConn(conn)
}

// FrameEnd 帧同步开始
func FrameEnd(args ...interface{}) {
	conn := args[iproto.Conn].(iduck.IConnection)
	if data := conn.GetNode(); data != nil {
		_node := data.(iduck.INode)
		_node.Complete()
	}
	_ = conn.Close()
}

// FrameMove move op
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
