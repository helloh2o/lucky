package inet

import (
	"errors"
	"github.com/google/uuid"
	"github.com/helloh2o/lucky/core/iduck"
	"github.com/helloh2o/lucky/core/iproto"
	"github.com/helloh2o/lucky/log"
	"sync/atomic"
	"time"
)

// FrameNode 帧同步节点
type FrameNode struct {
	// 节点ID
	NodeId string
	// 网络连接
	Connections map[interface{}]iduck.IConnection
	// 当前连接数量
	clientSize int64
	// 完成同步数量
	overSize int64
	// 同步周期
	FrameTicker *time.Ticker
	// current frame messages
	frameData [][]byte
	frameId   uint32
	allFrame  []interface{}
	// rand seed
	RandSeed int64
	// message channel
	onMessage chan []byte
	// AddConn
	addConnChan  chan iduck.IConnection
	delConnChan  chan string
	completeChan chan interface{}
	closeFlag    int64
}

// NewFrameNode return a new FrameNode
func NewFrameNode() *FrameNode {
	return &FrameNode{
		Connections:  make(map[interface{}]iduck.IConnection),
		NodeId:       uuid.New().String(),
		FrameTicker:  time.NewTicker(time.Millisecond * 66),
		RandSeed:     time.Now().UnixNano(),
		onMessage:    make(chan []byte),
		addConnChan:  make(chan iduck.IConnection),
		delConnChan:  make(chan string),
		completeChan: make(chan interface{}),
	}
}

// Serve the node
func (gr *FrameNode) Serve() {
	go func() {
		defer func() {
			for _, conn := range gr.Connections {
				conn.SetNode(nil)
			}
		}()
		for {
			// 优先管理连接状态
			select {
			// add conn
			case ic := <-gr.addConnChan:
				gr.Connections[ic.GetUuid()] = ic
				gr.clientSize++
				// conn leave
			case key := <-gr.delConnChan:
				delete(gr.Connections, key)
				gr.clientSize--
				// sync complete
			case <-gr.completeChan:
				gr.overSize++
				if gr.overSize >= gr.clientSize/2 {
					_ = gr.Destroy()
				}
			default:
				select {
				case <-gr.FrameTicker.C:
					gr.sendFrame()
				case pkg := <-gr.onMessage:
					if pkg == nil {
						log.Release("============= FrameNode %s, stop serve =============", gr.NodeId)
						// stop Serve
						gr.FrameTicker.Stop()
						return
					}
					gr.frameData = append(gr.frameData, pkg)
				}
			}
		}
	}()
}

func (gr *FrameNode) sendFrame() {
	// 没有消息
	if len(gr.frameData) == 0 || gr.clientSize == 0 {
		//log.Debug("Server empty frame without data")
		return
	}
	// 打包消息
	frame := iproto.ScFrame{
		Frame:     gr.frameId,
		Protocols: gr.frameData,
	}
	log.Debug("==> send frame to %d connections, contains %d package.", len(gr.Connections), len(gr.frameData))
	for _, conn := range gr.Connections {
		conn.WriteMsg(&frame)
	}
	// reset data
	gr.frameId++
	gr.frameData = gr.frameData[:0]
	gr.allFrame = append(gr.allFrame, gr.frameData)
}

// OnRawMessage msg
func (gr *FrameNode) OnRawMessage(msg []byte) error {
	if msg == nil {
		err := errors.New("can't frame nil message")
		return err
	}
	if gr.available() {
		gr.onMessage <- msg
	}
	return errFoo
}

// OnProtocolMessage interface
func (gr *FrameNode) OnProtocolMessage(interface{}) error {
	return nil
}

// GetAllMessage return chan []interface
func (gr *FrameNode) GetAllMessage() chan []interface{} {
	data := make(chan []interface{}, 1)
	data <- gr.allFrame
	return data
}

// AddConn conn
func (gr *FrameNode) AddConn(conn iduck.IConnection) error {
	if gr.available() {
		gr.addConnChan <- conn
		return nil
	}
	return errFoo
}

// DelConn by key
func (gr *FrameNode) DelConn(key string) error {
	if gr.available() {
		gr.delConnChan <- key
		return nil
	}
	return errFoo
}

// Complete sync
func (gr *FrameNode) Complete() error {
	if gr.available() {
		gr.completeChan <- struct{}{}
		return nil
	}
	return errFoo
}

// Destroy the node
func (gr *FrameNode) Destroy() error {
	if gr.available() {
		atomic.AddInt64(&gr.closeFlag, 1)
		go func() {
			gr.onMessage <- nil
		}()
		return nil
	}
	return errFoo
}

func (gr *FrameNode) available() bool {
	return atomic.LoadInt64(&gr.closeFlag) == 0
}
