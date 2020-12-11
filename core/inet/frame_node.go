package inet

import (
	"github.com/google/uuid"
	"lucky/core/iduck"
	"lucky/core/iproto"
	"lucky/log"
	"runtime/debug"
	"sync"
	"time"
)

// 帧同步节点
type FrameNode struct {
	// 网络连接
	Connections map[interface{}]iduck.IConnection
	// 当前连接数量
	clientSize int64
	// 完成同步数量
	overSize int64
	// 进入令牌
	EnterToken string
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
}

func NewFrameNode() *FrameNode {
	return &FrameNode{
		Connections:  make(map[interface{}]iduck.IConnection),
		EnterToken:   uuid.New().String(),
		FrameTicker:  time.NewTicker(time.Millisecond * 66),
		RandSeed:     time.Now().UnixNano(),
		onMessage:    make(chan []byte, 1000),
		addConnChan:  make(chan iduck.IConnection),
		delConnChan:  make(chan string),
		completeChan: make(chan interface{}),
	}
}

func (gr *FrameNode) Serve() {
	go func() {
		for {
			select {
			case <-gr.FrameTicker.C:
				gr.sendFrame()
			case pkg := <-gr.onMessage:
				if pkg == nil {
					log.Release("============= FrameNode %s, stop serve =============", gr.EnterToken)
					// stop Serve
					gr.FrameTicker.Stop()
					return
				}
				gr.frameData = append(gr.frameData, pkg)
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
					gr.Destroy()
				}
			}
		}
	}()
}

func (gr *FrameNode) sendFrame() {
	defer func() {
		if r := recover(); r != nil {
			log.Error("write frame error %v, stack %s", r, string(debug.Stack()))
			gr.Destroy()
		}
	}()
	// 没有消息
	if len(gr.frameData) == 0 {
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

func (gr *FrameNode) OnRawMessage(msg []byte) {
	select {
	case gr.onMessage <- msg:
	default:
	}
}

func (gr *FrameNode) OnProtocolMessage(interface{}) {}

func (gr *FrameNode) GetAllMessage() chan []interface{} {
	data := make(chan []interface{}, 1)
	data <- gr.allFrame
	return data
}

func (gr *FrameNode) AddConn(conn iduck.IConnection) {
	select {
	case gr.addConnChan <- conn:
	default:
	}
}

func (gr *FrameNode) DelConn(key string) {
	select {
	case gr.delConnChan <- key:
	default:
	}
}

// 完成同步
func (gr *FrameNode) Complete() {
	select {
	case gr.completeChan <- struct{}{}:
	default:
	}
}

// 摧毁节点
func (gr *FrameNode) Destroy() {
	var one sync.Once
	one.Do(func() {
		for _, conn := range gr.Connections {
			conn.SetNode(nil)
		}
		gr.onMessage <- nil
	})
}
