package lucky

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/helloh2o/lucky/log"
	"reflect"
)

// JsonProcessor one of Processor implement
type JsonProcessor struct {
	bigEndian bool
	enc       Encryptor
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
func (jp *JsonProcessor) OnReceivedPackage(writer interface{}, body []byte) error {
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
		return err
	}
	info, ok := jp.handlers[pack.Id]
	if !ok {
		log.Error("Not register msg id %d", pack.Id)
		return nil
	}
	msg := reflect.New(info.msgType.Elem()).Interface()
	msgBytes, _ := json.Marshal(pack.Content)
	err := json.Unmarshal(msgBytes, msg)
	if err != nil {
		log.Error("Can't unmarshal pack content to json msg, %+v", body)
		return err
	}
	// 执行逻辑
	execute(info, msg, writer, body, uint32(pack.Id))
	return nil
}

// WrapMsg format the interface message to []byte
func (jp *JsonProcessor) WrapMsg(message interface{}) ([]byte, error) {
	log.Debug("===> JSON processor warp %v for write", reflect.TypeOf(message))
	data, err := json.Marshal(message)
	if err != nil {
		return data, err
	}
	tp := reflect.TypeOf(message)
	id, ok := jp.msgTypes[tp]
	if !ok {
		return nil, fmt.Errorf("not register %v", tp)
	}
	protocol := JsonProtocol{
		Id:      id,
		Content: message,
	}
	data, err = json.Marshal(&protocol)
	if err != nil {
		return nil, err
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
	return pkg, nil
}

// WrapIdMsg format the interface message to []byte with id
func (jp *JsonProcessor) WrapIdMsg(id uint32, message interface{}) ([]byte, error) {
	log.Debug("===> JSON processor warp %v for write", reflect.TypeOf(message))
	data, err := json.Marshal(message)
	if err != nil {
		return data, err
	}
	protocol := JsonProtocol{
		Id:      int(id),
		Content: message,
	}
	data, err = json.Marshal(&protocol)
	if err != nil {
		return nil, err
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
	return pkg, nil
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
func (jp *JsonProcessor) SetEncryptor(enc Encryptor) {
	jp.enc = enc
}
