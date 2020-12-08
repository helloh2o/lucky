package iproto

import (
	"reflect"
)

type msgInfo struct {
	msgId       int
	msgType     reflect.Type
	msgCallback func(args ...interface{})
}
