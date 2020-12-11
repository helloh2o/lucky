package iproto

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"lucky/core/iduck"
	"lucky/log"
	"reflect"
	"runtime/debug"
)

type JsonProcessor struct {
	bigEndian bool
	enc       iduck.Encryptor
	msgTypes  map[reflect.Type]int
	handlers  map[int]msgInfo
}

// PB processor
func NewJSONProcessor() *JsonProcessor {
	pb := JsonProcessor{
		msgTypes: make(map[reflect.Type]int),
		handlers: make(map[int]msgInfo),
	}
	return &pb
}

// 收到完整数据包
func (jp *JsonProcessor) OnReceivedPackage(conn iduck.IConnection, body []byte) {
	// 解密
	if jp.enc != nil {
		//log.Debug("before decode:: %v", body)
		body = jp.enc.Decode(body)
		//log.Debug("after decode:: %v", body)
	}
	// 解码
	var pack Protocol
	if err := json.Unmarshal(body, &pack); err != nil {
		log.Error("Can't unmarshal pack body to json Protocol, %+v", body)
		return
	}
	h, ok := jp.handlers[int(pack.Id)]
	if !ok {
		log.Error("Not register msg id %d", pack.Id)
		return
	}
	msg := reflect.New(h.msgType.Elem()).Interface()
	err := json.Unmarshal(pack.Content, msg)
	if err != nil {
		log.Error("UnmarshalMerge pack.contents error by id %d", pack.Id)
		return
	}
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("%v", r)
				log.Error("panic at msg %d handler, stack %s", pack.Id, string(debug.Stack()))
			}
		}()
		h.msgCallback(msg, conn, body)
	}()
}

func (jp *JsonProcessor) WarpMsg(message interface{}) (error, []byte) {
	data, err := json.Marshal(message)
	if err != nil {
		return err, nil
	}
	tp := reflect.TypeOf(message)
	id, ok := jp.msgTypes[tp]
	if !ok {
		return errors.New(fmt.Sprintf("not register %v", tp)), nil
	}
	protocol := Protocol{
		Id:      uint32(id),
		Content: data,
	}
	data, err = json.Marshal(&protocol)
	if err != nil {
		return err, nil
	}
	if jp.enc != nil {
		//log.Debug("before encode:: %v", data)
		data = jp.enc.Encode(data)
		//log.Debug("after  encode:: %v", data)
	}
	// head
	head := make([]byte, 2)
	if jp.bigEndian {
		binary.BigEndian.PutUint16(head, uint16(len(data)))
	} else {
		binary.LittleEndian.PutUint16(head, uint16(len(data)))
	}
	pkg := append(head, data...)
	return nil, pkg
}

func (jp *JsonProcessor) RegisterHandler(id int, entity interface{}, handle func(args ...interface{})) {
	if _, ok := jp.handlers[id]; ok {
		log.Error("Already register handler by Id:: %d", id)
	} else {
		jp.handlers[id] = msgInfo{
			msgId:       id,
			msgType:     reflect.TypeOf(entity),
			msgCallback: handle,
		}
		jp.msgTypes[reflect.TypeOf(entity)] = id
	}
}

func (jp *JsonProcessor) SetBigEndian() {
	jp.bigEndian = true
}
func (jp *JsonProcessor) GetBigEndian() bool {
	return jp.bigEndian
}
func (jp *JsonProcessor) SetEncryptor(enc iduck.Encryptor) {
	jp.enc = enc
}
