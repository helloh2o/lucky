package inet

import (
	"github.com/google/uuid"
	"lucky/core/iduck"
	"lucky/log"
	"runtime/debug"
)

// 广播转发节点
type BroadcastNode struct {
	// 网络连接
	Connections map[interface{}]iduck.IConnection
	// 当前连接数量
	clientSize int64
	// 进入令牌
	NodeId string
	// message channel
	onMessage   chan interface{}
	allMessages []interface{}
	// AddConn
	addConnChan chan iduck.IConnection
	delConnChan chan string
}

func NewBroadcastNode() *BroadcastNode {
	return &BroadcastNode{
		Connections: make(map[interface{}]iduck.IConnection),
		NodeId:      uuid.New().String(),
		onMessage:   make(chan interface{}, 1000),
		addConnChan: make(chan iduck.IConnection),
		delConnChan: make(chan string),
	}
}

func (bNode *BroadcastNode) Serve() {
	go func() {
		for {
			select {
			case pkg := <-bNode.onMessage:
				if pkg == nil {
					log.Release("============= BroadcastNode %s, stop serve =============", bNode.NodeId)
					// stop Serve
					return
				}
				bNode.allMessages = append(bNode.allMessages, pkg)
				bNode.broadcast(pkg)
				// add conn
			case ic := <-bNode.addConnChan:
				bNode.Connections[ic.GetUuid()] = ic
				bNode.clientSize++
				// conn leave
			case key := <-bNode.delConnChan:
				delete(bNode.Connections, key)
				bNode.clientSize--
			}
		}
	}()
}

func (bNode *BroadcastNode) broadcast(msg interface{}) {
	if msg == nil {
		log.Error("can't broadcast nil msg.")
		return
	}
	defer func() {
		if r := recover(); r != nil {
			log.Error("write frame error %v, stack %s", r, string(debug.Stack()))
			bNode.Destroy()
		}
	}()
	for _, conn := range bNode.Connections {
		conn.WriteMsg(msg)
	}
}

func (bNode *BroadcastNode) OnRawMessage([]byte) {}

func (bNode *BroadcastNode) OnProtocolMessage(msg interface{}) {
	select {
	case bNode.onMessage <- msg:
	default:
	}
}

func (bNode *BroadcastNode) GetAllMessage() chan []interface{} {
	data := make(chan []interface{}, 1)
	data <- []interface{}{bNode.allMessages}
	return data
}

func (bNode *BroadcastNode) AddConn(conn iduck.IConnection) {
	select {
	case bNode.addConnChan <- conn:
	default:
	}
}

func (bNode *BroadcastNode) DelConn(key string) {
	select {
	case bNode.delConnChan <- key:
	default:
	}
}

// 完成同步
func (bNode *BroadcastNode) Complete() {
}

// 摧毁节点
func (bNode *BroadcastNode) Destroy() {
	bNode.onMessage <- nil
}
