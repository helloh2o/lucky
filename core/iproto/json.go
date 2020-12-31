package iproto

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/helloh2o/lucky/core/iduck"
	"github.com/helloh2o/lucky/log"
	"reflect"
)

// JsonProcessor one of Processor implement
type JsonProcessor struct {
	bigEndian bool
	enc       iduck.Encryptor
	msgTypes  map[reflect.Type]int
	handlers  map[int]msgInfo
}

// JsonProtocol is the protocol for JsonProcessor
type JsonProtocol struct {
	Id      int         `json:"id"`
	Content interface{} `json:"content"`
}

// NewJSONProcessor return new JsonProcessor
func NewJSONProcessor() *JsonProcessor {
	pb := JsonProcessor{
		msgTypes: make(map[reflect.Type]int),
		handlers: make(map[int]msgInfo),
	}
	return &pb
}

// OnReceivedPackage 收到完整数据包
func (jp *JsonProcessor) OnReceivedPackage(writer interface{}, body []byte) {
	// 解密
	if jp.enc != nil {
		//log.Debug("before decode:: %v", body)
		body = jp.enc.Decode(body)
		//log.Debug("after decode:: %v", body)
	}
	// 解码
	var pack JsonProtocol
	if err := json.Unmarshal(body, &pack); err != nil {
		log.Error("Can't unmarshal pack body to json Protocol, %+v", body)
		return
	}
	info, ok := jp.handlers[pack.Id]
	if !ok {
		log.Error("Not register msg id %d", pack.Id)
		return
	}
	msg := reflect.New(info.msgType.Elem()).Interface()
	msgBytes, _ := json.Marshal(pack.Content)
	err := json.Unmarshal(msgBytes, msg)
	if err != nil {
		log.Error("Can't unmarshal pack content to json msg, %+v", body)
		return
	}
	// 执行逻辑
	execute(info, msg, writer, body, uint32(pack.Id))
}

// WarpMsg format the interface message to []byte
func (jp *JsonProcessor) WarpMsg(message interface{}) (error, []byte) {
	data, err := json.Marshal(message)
	if err != nil {
		return err, data
	}
	tp := reflect.TypeOf(message)
	id, ok := jp.msgTypes[tp]
	if !ok {
		return errors.New(fmt.Sprintf("not register %v", tp)), nil
	}
	protocol := JsonProtocol{
		Id:      id,
		Content: message,
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

// RegisterHandler for logic
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

// SetBigEndian for order
func (jp *JsonProcessor) SetBigEndian() {
	jp.bigEndian = true
}

// GetBigEndian of the order
func (jp *JsonProcessor) GetBigEndian() bool {
	return jp.bigEndian
}

// SetEncryptor for processor
func (jp *JsonProcessor) SetEncryptor(enc iduck.Encryptor) {
	jp.enc = enc
}
