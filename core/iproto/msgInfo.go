package iproto

import (
	"github.com/helloh2o/lucky/log"
	"reflect"
	"runtime/debug"
	"time"
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
	begin := time.Now().UnixNano() / int64(time.Millisecond)
	mInfo.msgCallback(msg, writer, body)
	costs := time.Now().UnixNano()/int64(time.Millisecond) - begin
	log.Debug("===> execute logic %d costs %dms, msgType %v <===", mInfo.msgId, costs, mInfo.msgType)
}
