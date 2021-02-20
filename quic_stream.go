package lucky

import (
	"encoding/binary"
	"github.com/google/uuid"
	"github.com/helloh2o/lucky/log"
	"github.com/lucas-clemente/quic-go"
	"io"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

// QuicStream is warped udp conn for luck
type QuicStream struct {
	sync.RWMutex
	uuid string
	quic.Stream
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

// NewQuicStream return new udp conn
func NewQuicStream(stream quic.Stream, processor Processor) *QuicStream {
	if processor == nil || stream == nil {
		return nil
	}
	s := &QuicStream{
		uuid:       uuid.New().String(),
		Stream:     stream,
		writeQueue: make(chan []byte, C.ConnWriteQueueSize),
		processor:  processor,
		logicQueue: make(chan []byte, C.ConnUndoQueueSize),
	}
	// write q
	go func() {
		for pkg := range s.writeQueue {
			if pkg == nil {
				break
			}
			if C.ConnWriteTimeout > 0 {
				_ = s.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(C.ConnWriteTimeout)))
			}
			_, err := s.Write(pkg)
			if err != nil {
				log.Error("Quic Steam write %v", err)
				break
			}
			_ = s.SetWriteDeadline(time.Time{})
		}
		// write over or error
		_ = s.Close()
		log.Release("Stream %d <=> %s closed.", s.Stream.StreamID())
	}()
	// logic q
	go func() {
		for pkg := range s.logicQueue {
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
				s.processor.OnReceivedPackage(s, pkg)
			}()
		}
	}()
	return s
}

// GetUuid get uuid of conn
func (s *QuicStream) GetUuid() string {
	return s.uuid
}

// ReadMsg read | write end -> write | read end -> conn end
func (s *QuicStream) ReadMsg() {
	defer func() {
		s.logicQueue <- nil
		s.writeQueue <- nil
		// force close conn
		if !s.IsClosed() {
			_ = s.Close()
		}
	}()
	bf := make([]byte, C.MaxDataPackageSize)
	// 第一个包默认5秒
	timeout := time.Second * time.Duration(C.FirstPackageTimeout)
	for {
		_ = s.SetReadDeadline(time.Now().Add(timeout))
		// read length
		_, err := io.ReadAtLeast(s, bf[:2], 2)
		if err != nil {
			log.Error("Quic Steam read message head error %s", err.Error())
			return
		}
		var ln uint16
		if s.processor.GetBigEndian() {
			ln = binary.BigEndian.Uint16(bf[:2])
		} else {
			ln = binary.LittleEndian.Uint16(bf[:2])
		}
		if ln < 1 || int(ln) > C.MaxDataPackageSize {
			log.Error("Quic Steam message length %d invalid", ln)
			return
		}
		// read data
		_, err = io.ReadFull(s, bf[:ln])
		if err != nil {
			log.Error("Quic Steam read data err %s", err.Error())
			return
		}
		// clean
		_ = s.SetDeadline(time.Time{})
		// write to cache queue
		select {
		case s.logicQueue <- append(make([]byte, 0), bf[:ln]...):
		default:
			// ignore overflow package not close conn
			log.Error("Quic Steam %d logic queue overflow err, queue size %d", s.Stream.StreamID(), len(s.logicQueue))
		}
		// after first pack | check heartbeat
		timeout = time.Second * time.Duration(C.ConnReadTimeout)
	}
}

// WriteMsg warp msg base on conn's processor
func (s *QuicStream) WriteMsg(message interface{}) {
	pkg, err := s.processor.WrapMsg(message)
	if err != nil {
		log.Error("Quic Steam OnWrapMsg package error %s", err)
	} else {
	push:
		select {
		case s.writeQueue <- pkg:
		default:
			if s.IsClosed() {
				return
			}
			time.Sleep(time.Millisecond * 50)
			// re push
			goto push
		}
	}
}

// Close the conn
func (s *QuicStream) Close() error {
	s.Lock()
	defer func() {
		s.Unlock()
		// add close flag
		atomic.AddInt64(&s.closeFlag, 1)
		if s.closeCb != nil {
			s.closeCb()
		}
		// clean write q if not empty
		for len(s.writeQueue) > 0 {
			<-s.writeQueue
		}
	}()
	return s.Close()
}

// IsClosed return the status of conn
func (s *QuicStream) IsClosed() bool {
	return atomic.LoadInt64(&s.closeFlag) != 0
}

// AfterClose conn call back
func (s *QuicStream) AfterClose(cb func()) {
	s.Lock()
	defer s.Unlock()
	s.closeCb = cb
}

// SetData for conn
func (s *QuicStream) SetData(data interface{}) {
	s.Lock()
	defer s.Unlock()
	s.userData = data
}

// GetData from conn
func (s *QuicStream) GetData() interface{} {
	s.RLock()
	defer s.RUnlock()
	return s.userData
}

// SetNode for conn
func (s *QuicStream) SetNode(node INode) {
	s.Lock()
	defer s.Unlock()
	s.node = node
}

// GetNode from conn
func (s *QuicStream) GetNode() INode {
	s.RLock()
	defer s.RUnlock()
	return s.node
}
