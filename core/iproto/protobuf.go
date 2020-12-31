package iproto

import (
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/helloh2o/lucky/core/iduck"
	"github.com/helloh2o/lucky/log"
	"reflect"
)

// PbfProcessor one of Processor implement protoc --go_out=. *.proto
type PbfProcessor struct {
	bigEndian bool
	enc       iduck.Encryptor
	msgTypes  map[reflect.Type]int
	handlers  map[int]msgInfo
}

// NewPBProcessor return PB processor
func NewPBProcessor() *PbfProcessor {
	pb := PbfProcessor{
		msgTypes: make(map[reflect.Type]int),
		handlers: make(map[int]msgInfo),
	}
	return &pb
}

// OnReceivedPackage 收到完整数据包
func (pbf *PbfProcessor) OnReceivedPackage(writer interface{}, body []byte) {
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
	// 执行逻辑
	execute(info, msg, writer, body, pack.Id)
}

// WarpMsg format the interface message to []byte
func (pbf *PbfProcessor) WarpMsg(message interface{}) ([]byte, error) {
	data, err := proto.Marshal(message.(proto.Message))
	if err != nil {
		return nil, err
	}
	tp := reflect.TypeOf(message)
	id, ok := pbf.msgTypes[tp]
	if !ok {
		return nil, fmt.Errorf("not register %v", tp)
	}
	protocol := Protocol{
		Id:      uint32(id),
		Content: data,
	}
	data, err = proto.Marshal(&protocol)
	if err != nil {
		return nil, err
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
	return pkg, nil
}

// WarpMsgNoHeader without header length
func (pbf *PbfProcessor) WarpMsgNoHeader(message interface{}) ([]byte, error) {
	data, err := pbf.WarpMsg(message)
	if err != nil {
		return nil, err
	}
	return data[2:], nil
}

// RegisterHandler for logic
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

// SetBigEndian for order
func (pbf *PbfProcessor) SetBigEndian() {
	pbf.bigEndian = true
}

// GetBigEndian of the order
func (pbf *PbfProcessor) GetBigEndian() bool {
	return pbf.bigEndian
}

// SetEncryptor for processor
func (pbf *PbfProcessor) SetEncryptor(enc iduck.Encryptor) {
	pbf.enc = enc
}
