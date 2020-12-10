package inet

import (
	"github.com/gorilla/websocket"
	"lucky/conf"
	"lucky/core/iduck"
	"lucky/log"
	"runtime/debug"
	"time"
)

type WSConn struct {
	conn *websocket.Conn
	// 缓存队列
	writeQueue chan []byte
	readQueue  chan []byte
	// 消息处理器
	processor iduck.Processor
}

func NewWSConn(conn *websocket.Conn, processor iduck.Processor) *WSConn {
	if processor == nil || conn == nil {
		return nil
	}
	wc := &WSConn{
		conn:       conn,
		writeQueue: make(chan []byte, conf.C.ConnWriteQueueSize),
		processor:  processor,
		// 单个缓存100个为处理的包
		readQueue: make(chan []byte, conf.C.ConnUndoQueueSize),
	}
	// write q
	go func() {
		for pkg := range wc.writeQueue {
			// read over
			if pkg == nil {
				break
			}
			// Binary=1 Text=0
			err := wc.conn.WriteMessage(websocket.BinaryMessage, pkg)
			if err != nil {
				log.Error("websocket write %v", err)
				break
			}
		}
		// write over or error
		_ = conn.Close()
		log.Release("Conn %s <=> %s closed.", wc.conn.LocalAddr(), wc.conn.RemoteAddr())
	}()
	// read q
	go func() {
		for pkg := range wc.readQueue {
			// read over
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

func (wc *WSConn) ReadMsg() {
	defer func() {
		wc.readQueue <- nil
		wc.writeQueue <- nil
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
			case wc.readQueue <- body:
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
		select {
		// ws write data only ,not need data length
		case wc.writeQueue <- pkg[2:]:
		default:
			log.Error(" =============== Drop message, write chan is full  %d  =============== ", len(wc.writeQueue))
		}
	}
}

func (wc *WSConn) Close() error {
	return wc.conn.Close()
}
