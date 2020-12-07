package iproto

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
	"lucky-day/core/duck"
	"reflect"
)

/*
[msgId: other protocol]
protocol := protocolMsg{
MsgId:    id,
Contents: data,
}*/

type PbfProcessor struct {
	big      bool
	enc      duck.Encrypt
	msgTypes map[reflect.Type]int
	handlers map[int]msgInfo
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
		logrus.Error("Can't unmarshal pack body to protocolMsg, %+v", body)
		return
	}
	info, ok := pbf.handlers[int(pack.Id)]
	if !ok {
		logrus.Error("Not register msg id %d", pack.Id)
		return
	}
	msg := reflect.New(info.msgType.Elem()).Interface()
	err := proto.UnmarshalMerge(pack.Content, msg.(proto.Message))
	if err != nil {
		logrus.Error("UnmarshalMerge pack.contents error by id %d", pack.Id)
		return
	}
	// route
	info.msgCallback(msg, conn)
}

func (pbf *PbfProcessor) OnWarpMsg(message interface{}) (error, []byte) {
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
	if pbf.big {
		binary.BigEndian.PutUint16(head, uint16(len(data)))
	} else {
		binary.LittleEndian.PutUint16(head, uint16(len(data)))
	}
	pkg := append(head, data...)
	return nil, pkg
}

func (pbf *PbfProcessor) RegisterHandler(id int, entity interface{}, handle func(args ...interface{})) {
	if _, ok := pbf.handlers[id]; ok {
		logrus.Error("Already register handler by Id:: %d", id)
	} else {
		pbf.handlers[id] = msgInfo{
			msgId:       id,
			msgType:     reflect.TypeOf(entity),
			msgCallback: handle,
		}
		pbf.msgTypes[reflect.TypeOf(entity)] = id
	}
}

func (pbf *PbfProcessor) SetBytesOrder(big bool) {
	pbf.big = big
}
func (pbf *PbfProcessor) GetBigOrder() bool {
	return pbf.big
}
func (pbf *PbfProcessor) SetEncrypt(enc duck.Encrypt) {
	pbf.enc = enc
}
func (pbf *PbfProcessor) GetEncrypt() duck.Encrypt {
	return pbf.enc
}
