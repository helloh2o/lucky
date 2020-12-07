package iproto

import (
	"github.com/golang/protobuf/proto"
	"reflect"
)

type msgInfo struct {
	msgId       int
	msgType     reflect.Type
	msgCallback func(args ...interface{})
}

type protocolMsg struct {
	proto.Message
	MsgId    int
	Contents []byte
}
