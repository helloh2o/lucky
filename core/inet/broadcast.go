package inet

import (
	"errors"
	"github.com/google/uuid"
	"github.com/helloh2o/lucky/core/iduck"
	"github.com/helloh2o/lucky/log"
	"runtime/debug"
	"sync/atomic"
	"time"
)

// BroadcastNode 广播转发节点
type BroadcastNode struct {
	// 节点ID
	NodeId string
	// 网络连接
	Connections map[interface{}]iduck.IConnection
	// 当前连接数量
	clientSize int64
	// message channel
	onMessage      chan interface{}
	recentMessages []interface{}
	// AddConn
	addConnChan chan iduck.IConnection
	delConnChan chan string
	closeFlag   int64
}

// closedErr node closed error
var closedErr = errors.New("broadcast node closed")

func NewBroadcastNode() *BroadcastNode {
	return &BroadcastNode{
		Connections: make(map[interface{}]iduck.IConnection),
		NodeId:      uuid.New().String(),
		onMessage:   make(chan interface{}),
		addConnChan: make(chan iduck.IConnection),
		delConnChan: make(chan string),
	}
}

// Serve the node
func (bNode *BroadcastNode) Serve() {
	go func() {
		defer func() {
			for _, conn := range bNode.Connections {
				conn.SetNode(nil)
			}
		}()
		for {
			// 优先管理连接
			select {
			// add conn
			case ic := <-bNode.addConnChan:
				bNode.Connections[ic.GetUuid()] = ic
				bNode.clientSize++
			// conn leave
			case key := <-bNode.delConnChan:
				delete(bNode.Connections, key)
				bNode.clientSize--
			default:
				select {
				case pkg := <-bNode.onMessage:
					if pkg == nil {
						log.Release("============= BroadcastNode %s, stop serve =============", bNode.NodeId)
						// stop Serve
						return
					}
					bNode.recentMessages = append(bNode.recentMessages, pkg)
					// cache recent 100
					recentSize := len(bNode.recentMessages)
					if recentSize > 100 {
						bNode.recentMessages = bNode.recentMessages[recentSize-100:]
					}
					bNode.broadcast(pkg)
				default:
					time.Sleep(time.Millisecond * 50)
				}
			}
		}
	}()
}

func (bNode *BroadcastNode) broadcast(msg interface{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("write frame error %v, stack %s", r, string(debug.Stack()))
		}
	}()
	if bNode.clientSize == 0 {
		return
	}
	log.Debug("amount %d, broadcast msg %+v", bNode.clientSize, msg)
	for _, conn := range bNode.Connections {
		conn.WriteMsg(msg)
	}
	log.Debug(" ======= broadcast ok ======= ")
}

// OnRawMessage bytes
func (bNode *BroadcastNode) OnRawMessage([]byte) error { return nil }

// OnProtocolMessage interface
func (bNode *BroadcastNode) OnProtocolMessage(msg interface{}) error {
	if bNode.available() {
		bNode.onMessage <- msg
	}
	return closedErr
}

// GetAllMessage return  chan []interface{}
func (bNode *BroadcastNode) GetAllMessage() chan []interface{} {
	data := make(chan []interface{}, 1)
	data <- bNode.recentMessages
	return data
}

// AddConn by conn
func (bNode *BroadcastNode) AddConn(conn iduck.IConnection) error {
	if bNode.available() {
		bNode.addConnChan <- conn
		return nil
	}
	return closedErr
}

// DelConn by key
func (bNode *BroadcastNode) DelConn(key string) error {
	if bNode.available() {
		bNode.delConnChan <- key
		return nil
	}
	return closedErr
}

// Complete sync
func (bNode *BroadcastNode) Complete() error {
	return nil
}

// Destroy the node
func (bNode *BroadcastNode) Destroy() error {
	if bNode.available() {
		atomic.AddInt64(&bNode.closeFlag, 1)
		go func() {
			bNode.onMessage <- nil
		}()
	}
	return closedErr
}

func (bNode *BroadcastNode) available() bool {
	return atomic.LoadInt64(&bNode.closeFlag) == 0
}
