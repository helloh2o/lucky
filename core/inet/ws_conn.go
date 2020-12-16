package inet

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"lucky/conf"
	"lucky/core/iduck"
	"lucky/log"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

type WSConn struct {
	sync.RWMutex
	uuid string
	conn *websocket.Conn
	// 缓写队列
	writeQueue chan []byte
	// 逻辑消息队列
	logicQueue chan []byte
	// 消息处理器
	processor iduck.Processor
	userData  interface{}
	node      iduck.INode
	// after close
	closeCb   func()
	closeFlag int64
}

func NewWSConn(conn *websocket.Conn, processor iduck.Processor) *WSConn {
	if processor == nil || conn == nil {
		return nil
	}
	wc := &WSConn{
		uuid:       uuid.New().String(),
		conn:       conn,
		writeQueue: make(chan []byte, conf.C.ConnWriteQueueSize),
		processor:  processor,
		// 单个缓存100个为处理的包
		logicQueue: make(chan []byte, conf.C.ConnUndoQueueSize),
	}
	// write q
	go func() {
		for pkg := range wc.writeQueue {
			if pkg == nil {
				break
			}
			// Binary=1 Text=0
			if conf.C.ConnWriteTimeout > 0 {
				_ = wc.conn.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(conf.C.ConnWriteTimeout)))
			}
			err := wc.conn.WriteMessage(websocket.BinaryMessage, pkg)
			if err != nil {
				log.Error("websocket write %v", err)
				break
			}
			_ = wc.conn.SetWriteDeadline(time.Time{})
		}
		// write over or error
		_ = wc.Close()
		log.Release("Conn %s <=> %s closed.", wc.conn.LocalAddr(), wc.conn.RemoteAddr())
	}()
	// logic q
	go func() {
		for pkg := range wc.logicQueue {
			// logic over
			if pkg == nil {
				break
			}
			// processor handle the package
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Error("panic %v in processor, stack %s", r, string(debug.Stack()))
					}
				}()
				wc.processor.OnReceivedPackage(wc, pkg)
			}()
		}
	}()
	return wc
}

func (wc *WSConn) GetUuid() string {
	return wc.uuid
}

// read | write end -> write | read end -> conn end
func (wc *WSConn) ReadMsg() {
	defer func() {
		wc.logicQueue <- nil
		wc.writeQueue <- nil
		// force close conn
		if !wc.IsClosed() {
			_ = wc.conn.Close()
		}
	}()
	timeout := time.Second * time.Duration(conf.C.FirstPackageTimeout)
	for {
		_ = wc.conn.SetReadDeadline(time.Now().Add(timeout))
		typee, body, err := wc.conn.ReadMessage()
		if err != nil {
			break
		}
		switch typee {
		case websocket.BinaryMessage:
			// write to cache queue
			select {
			case wc.logicQueue <- body:
			default:
				log.Error("WSConn read queue overflow err %v", err)
				return
			}
			// clean
			_ = wc.conn.SetReadDeadline(time.Time{})
			timeout = time.Second * time.Duration(conf.C.ConnReadTimeout)
		case websocket.TextMessage:
			log.Error("not support pain text message type %d", typee)
			return
		}
	}
}

func (wc *WSConn) WriteMsg(message interface{}) {
	err, pkg := wc.processor.WarpMsg(message)
	if err != nil {
		log.Error("OnWarpMsg package error %s", err)
	} else {
		// ws write data only ,not need data length
	push:
		select {
		case wc.writeQueue <- pkg[2:]:
		default:
			if wc.IsClosed() {
				return
			}
			time.Sleep(time.Millisecond * 50)
			// re push
			goto push
		}

	}
}

func (wc *WSConn) Close() error {
	wc.Lock()
	defer func() {
		wc.Unlock()
		// add closed flag
		atomic.AddInt64(&wc.closeFlag, 1)
		if wc.closeCb != nil {
			wc.closeCb()
		}
		// clean write q if not empty
		for len(wc.writeQueue) > 0 {
			<-wc.writeQueue
		}
	}()
	return wc.conn.Close()
}

func (wc *WSConn) IsClosed() bool {
	return atomic.LoadInt64(&wc.closeFlag) != 0
}

func (wc *WSConn) AfterClose(cb func()) {
	wc.Lock()
	defer wc.Unlock()
	wc.closeCb = cb
}
func (wc *WSConn) SetData(data interface{}) {
	wc.Lock()
	defer wc.Unlock()
	wc.userData = data
}
func (wc *WSConn) GetData() interface{} {
	wc.RLock()
	defer wc.RUnlock()
	return wc.userData
}
func (wc *WSConn) SetNode(node iduck.INode) {
	wc.Lock()
	defer wc.Unlock()
	wc.node = node
}
func (wc *WSConn) GetNode() iduck.INode {
	wc.RLock()
	defer wc.RUnlock()
	return wc.node
}
