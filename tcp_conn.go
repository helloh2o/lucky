package lucky

import (
	"encoding/binary"
	"github.com/google/uuid"
	"github.com/helloh2o/lucky/log"
	"io"
	"net"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

// TCPConn is warped tcp conn for luck
type TCPConn struct {
	sync.RWMutex
	uuid string
	net.Conn
	// 缓写队列
	writeQueue chan []byte
	// 逻辑消息队列
	logicQueue chan []byte
	// 消息处理器
	processor Processor
	userData  interface{}
	node      INode
	// after close
	closeCb   func()
	closeFlag int64
}

// KCPConn 可靠的UDP，like TCP
type KCPConn struct {
	*TCPConn
}

// NewKcpConn get new kcp conn
func NewKcpConn(conn net.Conn, processor Processor) *KCPConn {
	tcpConn := NewTcpConn(conn, processor)
	if tcpConn != nil {
		return &KCPConn{tcpConn}
	}
	return nil
}

// NewTcpConn return new tcp conn
func NewTcpConn(conn net.Conn, processor Processor) *TCPConn {
	if processor == nil || conn == nil {
		return nil
	}
	tc := &TCPConn{
		uuid:       uuid.New().String(),
		Conn:       conn,
		writeQueue: make(chan []byte, C.ConnWriteQueueSize),
		processor:  processor,
		// 单个缓存100个为处理的包
		logicQueue: make(chan []byte, C.ConnUndoQueueSize),
	}
	// write q
	go func() {
		for pkg := range tc.writeQueue {
			if pkg == nil {
				break
			}
			if C.ConnWriteTimeout > 0 {
				_ = tc.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(C.ConnWriteTimeout)))
			}
			_, err := tc.Write(pkg)
			if err != nil {
				log.Error("tcp write %v", err)
				break
			}
			_ = tc.SetWriteDeadline(time.Time{})
		}
		// write over or error
		_ = tc.Close()
		log.Release("Conn %s <=> %s closed.", tc.Conn.LocalAddr(), tc.Conn.RemoteAddr())
	}()
	// logic q
	go func() {
		for pkg := range tc.logicQueue {
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
				tc.processor.OnReceivedPackage(tc, pkg)
			}()
		}
	}()
	return tc
}

// GetUuid get uuid of conn
func (tc *TCPConn) GetUuid() string {
	return tc.uuid
}

// ReadMsg read | write end -> write | read end -> conn end
func (tc *TCPConn) ReadMsg() {
	defer func() {
		tc.logicQueue <- nil
		tc.writeQueue <- nil
		// force close conn
		if !tc.IsClosed() {
			_ = tc.Close()
		}
	}()
	bf := make([]byte, C.MaxDataPackageSize)
	// 第一个包默认5秒
	timeout := time.Second * time.Duration(C.FirstPackageTimeout)
	for {
		_ = tc.SetReadDeadline(time.Now().Add(timeout))
		// read length
		_, err := io.ReadAtLeast(tc, bf[:2], 2)
		if err != nil {
			log.Error("TCPConn read message head error %s", err.Error())
			return
		}
		var ln uint16
		if tc.processor.GetBigEndian() {
			ln = binary.BigEndian.Uint16(bf[:2])
		} else {
			ln = binary.LittleEndian.Uint16(bf[:2])
		}
		if ln < 1 || int(ln) > C.MaxDataPackageSize {
			log.Error("TCPConn message length %d invalid", ln)
			return
		}
		// read data
		_, err = io.ReadFull(tc, bf[:ln])
		if err != nil {
			log.Error("TCPConn read data err %s", err.Error())
			return
		}
		// clean
		_ = tc.SetDeadline(time.Time{})
		// write to cache queue
		select {
		case tc.logicQueue <- append(make([]byte, 0), bf[:ln]...):
		default:
			// ignore overflow package not close conn
			log.Error("TCPConn %s <=> %s logic queue overflow err, queue size %d", tc.LocalAddr(), tc.RemoteAddr(), len(tc.logicQueue))
		}
		// after first pack | check heartbeat
		timeout = time.Second * time.Duration(C.ConnReadTimeout)
	}
}

// WriteMsg warp msg base on conn's processor
func (tc *TCPConn) WriteMsg(message interface{}) {
	pkg, err := tc.processor.WrapMsg(message)
	if err != nil {
		log.Error("OnWrapMsg package error %s", err)
	} else {
	push:
		select {
		case tc.writeQueue <- pkg:
		default:
			if tc.IsClosed() {
				return
			}
			time.Sleep(time.Millisecond * 50)
			// re push
			goto push
		}
	}
}

// Close the conn
func (tc *TCPConn) Close() error {
	tc.Lock()
	defer func() {
		tc.Unlock()
		// add close flag
		atomic.AddInt64(&tc.closeFlag, 1)
		if tc.closeCb != nil {
			tc.closeCb()
		}
		// clean write q if not empty
		for len(tc.writeQueue) > 0 {
			<-tc.writeQueue
		}
	}()
	return tc.Conn.Close()
}

// IsClosed return the status of conn
func (tc *TCPConn) IsClosed() bool {
	return atomic.LoadInt64(&tc.closeFlag) != 0
}

// AfterClose conn call back
func (tc *TCPConn) AfterClose(cb func()) {
	tc.Lock()
	defer tc.Unlock()
	tc.closeCb = cb
}

// SetData for conn
func (tc *TCPConn) SetData(data interface{}) {
	tc.Lock()
	defer tc.Unlock()
	tc.userData = data
}

// GetData from conn
func (tc *TCPConn) GetData() interface{} {
	tc.RLock()
	defer tc.RUnlock()
	return tc.userData
}

// SetNode for conn
func (tc *TCPConn) SetNode(node INode) {
	tc.Lock()
	defer tc.Unlock()
	tc.node = node
}

// GetNode from conn
func (tc *TCPConn) GetNode() INode {
	tc.RLock()
	defer tc.RUnlock()
	return tc.node
}
