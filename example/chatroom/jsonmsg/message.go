package jsonmsg

import (
	"lucky/core/iduck"
	"lucky/core/iproto"
	"lucky/example/chatroom/chatnode"
	"lucky/log"
)

const (
	Enter_Room   = 1001
	Chat_Message = 1002
	Leave_Room   = 1003

	Join_Success = 2001
)

type EnterRoom struct {
}

type JoinRoomSuccess struct {
}

type ChatMessage struct {
	FromName string
	Content  string
}

type LeaveRoom struct {
}

var Processor = iproto.NewJSONProcessor()

func init() {
	Processor.RegisterHandler(Enter_Room, &EnterRoom{}, func(args ...interface{}) {
		conn := args[iproto.Conn].(iduck.IConnection)
		conn.AfterClose(func() {
			chatnode.GetRoom().DelConn(conn.GetUuid())
		})
		chatnode.GetRoom().AddConn(conn)
		conn.SetNode(chatnode.GetRoom())
		conn.WriteMsg(&JoinRoomSuccess{})
		// 房间的最近20条历史消息
		msgs := <-chatnode.GetRoom().GetAllMessage()
		var record []interface{}
		if len(msgs) > 20 {
			record = append(record, msgs[:20]...)
		} else {
			record = msgs
			for _, m := range record {
				conn.WriteMsg(m)
			}
			log.Debug("write %d history message.", len(record))
		}
	})

	// 将聊天消息转发给节点
	Processor.RegisterHandler(Chat_Message, &ChatMessage{}, func(args ...interface{}) {
		conn := args[iproto.Conn].(iduck.IConnection)
		if nd := conn.GetNode(); nd != nil {
			nd.OnProtocolMessage(args[iproto.Msg].(*ChatMessage))
		}
	})

	Processor.RegisterHandler(Leave_Room, &LeaveRoom{}, func(args ...interface{}) {
		conn := args[iproto.Conn].(iduck.IConnection)
		if nd := conn.GetNode(); nd != nil {
			nd.DelConn(conn.GetUuid())
		}
	})

	Processor.RegisterHandler(Join_Success, &JoinRoomSuccess{}, nil)
}
