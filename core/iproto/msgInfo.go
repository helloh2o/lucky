package iproto

import (
	"lucky/log"
	"reflect"
	"runtime/debug"
)

// 回调传参常量
const (
	Msg = iota
	Conn
	Raw
)

// 消息信息
type msgInfo struct {
	msgId       int
	msgType     reflect.Type
	msgCallback func(args ...interface{})
}

// 执行消息回调
func execute(mInfo msgInfo, msg interface{}, writer interface{}, body []byte, id uint32) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("%v", r)
			log.Error("panic at msg %d handler, stack %s", id, string(debug.Stack()))
		}
	}()
	mInfo.msgCallback(msg, writer, body)
}
