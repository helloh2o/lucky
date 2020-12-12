package iproto

import (
	"reflect"
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
