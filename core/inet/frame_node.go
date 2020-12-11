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
	sync.RWMutex
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
	allFrame  [][][]byte
	// rand seed
	RandSeed int64
	// message channel
	onMessage chan []byte
}

func NewFrameNode() *FrameNode {
	return &FrameNode{
		Connections: make(map[interface{}]iduck.IConnection),
		EnterToken:  uuid.New().String(),
		FrameTicker: time.NewTicker(time.Millisecond * 66),
		RandSeed:    time.Now().UnixNano(),
		onMessage:   make(chan []byte, 1000),
	}
}

func (gr *FrameNode) Serve() {
	go func() {
		for {
			select {
			case <-gr.FrameTicker.C:
				gr.sendFrame()
			default:
				select {
				default:
					continue
				case pkg := <-gr.onMessage:
					if pkg == nil {
						log.Release("============= Node %s, stop serve =============", gr.EnterToken)
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
	defer func() {
		if r := recover(); r != nil {
			log.Error("write frame error %v, stack %s", r, string(debug.Stack()))
			gr.Destroy()
		}
	}()
	// 没有消息
	if len(gr.frameData) == 0 {
		log.Debug("Server empty frame without data")
		return
	}
	gr.RLock()
	defer gr.RUnlock()
	// 打包消息
	frame := iproto.ScFrame{
		Frame:     gr.frameId + 1,
		Protocols: gr.frameData,
	}
	log.Debug("begin to send frame to %d connections, contains %d package.", len(gr.Connections), len(gr.frameData))
	for _, conn := range gr.Connections {
		conn.WriteMsg(&frame)
	}
	// reset data
	gr.frameId++
	gr.frameData = gr.frameData[:0]
	gr.allFrame = append(gr.allFrame, gr.frameData)
}

func (gr *FrameNode) OnMessage(msg []byte) {
	gr.onMessage <- msg
}

func (gr *FrameNode) GetAllMessage() [][][]byte {
	gr.RLock()
	defer gr.RUnlock()
	return gr.allFrame
}

func (gr *FrameNode) AddConn(key interface{}, conn iduck.IConnection) {
	gr.Lock()
	defer gr.Unlock()
	gr.Connections[key] = conn
	gr.clientSize++
}

func (gr *FrameNode) DelConn(key interface{}) {
	gr.Lock()
	defer gr.Unlock()
	delete(gr.Connections, key)
	gr.clientSize--
}

// 完成同步
func (gr *FrameNode) Complete() {
	gr.Lock()
	defer gr.Unlock()
	gr.overSize++
	if gr.overSize >= gr.clientSize/2 {
		gr.Destroy()
	}
}

// 摧毁节点
func (gr *FrameNode) Destroy() {
	gr.Lock()
	defer gr.Unlock()
	var one sync.Once
	one.Do(func() {
		for _, conn := range gr.Connections {
			conn.SetNode(nil)
		}
		gr.onMessage <- nil
	})
}
