package iproto

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"lucky-day/core/duck"
	"lucky-day/log"
	"reflect"
)

/*
[msgId: other protocol]
protocol := protocolMsg{
MsgId:    id,
Contents: data,
}*/

type PbfProcessor struct {
	bigEndian bool
	enc       duck.Encrypt
	msgTypes  map[reflect.Type]int
	handlers  map[int]msgInfo
}

// PB processor
func NewPBProcessor() *PbfProcessor {
	pbp := PbfProcessor{
		msgTypes: make(map[reflect.Type]int),
		handlers: make(map[int]msgInfo),
	}
	return &pbp
}

// do message
func (pbf *PbfProcessor) OnReceivedMsg(conn duck.IConnection, body []byte) {
	// 解密
	if pbf.enc != nil {
		pbf.enc.Decode(body)
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
	// route
	info.msgCallback(msg, conn)
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

func (pbf *PbfProcessor) SetBigEndian(big bool) {
	pbf.bigEndian = big
}
func (pbf *PbfProcessor) GetBigEndian() bool {
	return pbf.bigEndian
}
func (pbf *PbfProcessor) SetEncrypt(enc duck.Encrypt) {
	pbf.enc = enc
}
