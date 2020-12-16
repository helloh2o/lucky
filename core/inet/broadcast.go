package inet

import (
	"errors"
	"github.com/google/uuid"
	"lucky/core/iduck"
	"lucky/log"
	"runtime/debug"
	"sync/atomic"
	"time"
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
	closeFlag   int64
}

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
					bNode.allMessages = append(bNode.allMessages, pkg)
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
	log.Debug("amount %d, broadcast msg %+v", bNode.clientSize, msg)
	for _, conn := range bNode.Connections {
		conn.WriteMsg(msg)
	}
}

func (bNode *BroadcastNode) OnRawMessage([]byte) error { return nil }

func (bNode *BroadcastNode) OnProtocolMessage(msg interface{}) error {
	if bNode.available() {
		bNode.onMessage <- msg
	}
	return closedErr
}

func (bNode *BroadcastNode) GetAllMessage() chan []interface{} {
	data := make(chan []interface{}, 1)
	data <- bNode.allMessages
	return data
}

func (bNode *BroadcastNode) AddConn(conn iduck.IConnection) error {
	if bNode.available() {
		bNode.addConnChan <- conn
		return nil
	}
	return closedErr
}

func (bNode *BroadcastNode) DelConn(key string) error {
	if bNode.available() {
		bNode.delConnChan <- key
		return nil
	}
	return closedErr
}

// 完成同步
func (bNode *BroadcastNode) Complete() error {
	return nil
}

// 摧毁节点
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
