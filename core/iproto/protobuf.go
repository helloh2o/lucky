package iproto

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"lucky/core/iduck"
	"lucky/log"
	"reflect"
	"runtime/debug"
)

/*
[msgId: other protocol]
protocol := protocolMsg{
MsgId:    id,
Contents: data,
}*/
// protoc --go_out=. *.proto
type PbfProcessor struct {
	bigEndian bool
	enc       iduck.Encryptor
	msgTypes  map[reflect.Type]int
	handlers  map[int]msgInfo
}

// PB processor
func NewPBProcessor() *PbfProcessor {
	pb := PbfProcessor{
		msgTypes: make(map[reflect.Type]int),
		handlers: make(map[int]msgInfo),
	}
	return &pb
}

// 收到完整数据包
func (pbf *PbfProcessor) OnReceivedPackage(conn iduck.IConnection, body []byte) {
	// 如果连接在某个同步节点上，转发消息到节点
	if conn.GetNode() != nil {
		conn.GetNode().OnMessage(body)
		return
	}
	// 解密
	if pbf.enc != nil {
		//log.Debug("before decode:: %v", body)
		body = pbf.enc.Decode(body)
		//log.Debug("after decode:: %v", body)
	}
	// 解码
	var pack Protocol
	if err := proto.UnmarshalMerge(body, &pack); err != nil {
		log.Error("Can't unmarshal pack body to protocolMsg, %+v", body)
		return
	}
	info, ok := pbf.handlers[int(pack.Id)]
	if !ok {
		log.Error("Not register msg id %d", pack.Id)
		return
	}
	msg := reflect.New(info.msgType.Elem()).Interface()
	err := proto.UnmarshalMerge(pack.Content, msg.(proto.Message))
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
		info.msgCallback(msg, conn)
	}()
}

func (pbf *PbfProcessor) WarpMsg(message interface{}) (error, []byte) {
	data, err := proto.Marshal(message.(proto.Message))
	if err != nil {
		return err, nil
	}
	tp := reflect.TypeOf(message)
	id, ok := pbf.msgTypes[tp]
	if !ok {
		return errors.New(fmt.Sprintf("not register %v", tp)), nil
	}
	protocol := Protocol{
		Id:      uint32(id),
		Content: data,
	}
	data, err = proto.Marshal(&protocol)
	if err != nil {
		return err, nil
	}
	if pbf.enc != nil {
		//log.Debug("before encode:: %v", data)
		data = pbf.enc.Encode(data)
		//log.Debug("after  encode:: %v", data)
	}
	// head
	head := make([]byte, 2)
	if pbf.bigEndian {
		binary.BigEndian.PutUint16(head, uint16(len(data)))
	} else {
		binary.LittleEndian.PutUint16(head, uint16(len(data)))
	}
	pkg := append(head, data...)
	return nil, pkg
}

func (pbf *PbfProcessor) RegisterHandler(id int, entity interface{}, handle func(args ...interface{})) {
	if _, ok := pbf.handlers[id]; ok {
		log.Error("Already register handler by Id:: %d", id)
	} else {
		pbf.handlers[id] = msgInfo{
			msgId:       id,
			msgType:     reflect.TypeOf(entity),
			msgCallback: handle,
		}
		pbf.msgTypes[reflect.TypeOf(entity)] = id
	}
}

func (pbf *PbfProcessor) SetBigEndian() {
	pbf.bigEndian = true
}
func (pbf *PbfProcessor) GetBigEndian() bool {
	return pbf.bigEndian
}
func (pbf *PbfProcessor) SetEncryptor(enc iduck.Encryptor) {
	pbf.enc = enc
}
