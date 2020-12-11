package inet

import (
	"github.com/google/uuid"
	"lucky/core/iduck"
	"lucky/log"
	"runtime/debug"
	"sync"
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
	addConnChan  chan iduck.IConnection
	delConnChan  chan string
	completeChan chan interface{}
}

func NewBroadcastNode() *BroadcastNode {
	return &BroadcastNode{
		Connections:  make(map[interface{}]iduck.IConnection),
		NodeId:       uuid.New().String(),
		onMessage:    make(chan interface{}, 1000),
		addConnChan:  make(chan iduck.IConnection),
		delConnChan:  make(chan string),
		completeChan: make(chan interface{}),
	}
}

func (bnode *BroadcastNode) Serve() {
	go func() {
		for {
			select {
			case pkg := <-bnode.onMessage:
				if pkg == nil {
					log.Release("============= BroadcastNode %s, stop serve =============", bnode.NodeId)
					// stop Serve
					return
				}
				bnode.allMessages = append(bnode.allMessages, pkg)
				bnode.broadcast(pkg)
				// add conn
			case ic := <-bnode.addConnChan:
				bnode.Connections[ic.GetUuid()] = ic
				bnode.clientSize++
				// conn leave
			case key := <-bnode.delConnChan:
				delete(bnode.Connections, key)
				bnode.clientSize--
			}
		}
	}()
}

func (bnode *BroadcastNode) broadcast(msg interface{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("write frame error %v, stack %s", r, string(debug.Stack()))
			bnode.Destroy()
		}
	}()
	for _, conn := range bnode.Connections {
		conn.WriteMsg(msg)
	}
}

func (bnode *BroadcastNode) OnRawMessage([]byte) {}

func (bnode *BroadcastNode) OnProtocolMessage(msg interface{}) {
	select {
	case bnode.onMessage <- msg:
	default:
	}
}

func (bnode *BroadcastNode) GetAllMessage() chan []interface{} {
	data := make(chan []interface{}, 1)
	data <- []interface{}{bnode.allMessages}
	return data
}

func (bnode *BroadcastNode) AddConn(conn iduck.IConnection) {
	select {
	case bnode.addConnChan <- conn:
	default:
	}
}

func (bnode *BroadcastNode) DelConn(key string) {
	select {
	case bnode.delConnChan <- key:
	default:
	}
}

// 完成同步
func (bnode *BroadcastNode) Complete() {
	select {
	case bnode.completeChan <- struct{}{}:
	default:
	}
}

// 摧毁节点
func (bnode *BroadcastNode) Destroy() {
	var one sync.Once
	one.Do(func() {
		for _, conn := range bnode.Connections {
			conn.SetNode(nil)
		}
		bnode.onMessage <- nil
	})
}
