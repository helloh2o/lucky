package inet

import (
	"encoding/binary"
	"github.com/google/uuid"
	"github.com/lucas-clemente/quic-go"
	"io"
	"lucky/conf"
	"lucky/core/iduck"
	"lucky/log"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

type QuicSteam struct {
	sync.RWMutex
	uuid string
	quic.Stream
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

func NewQuicSteam(stream quic.Stream, processor iduck.Processor) *QuicSteam {
	if processor == nil || stream == nil {
		return nil
	}
	s := &QuicSteam{
		uuid:       uuid.New().String(),
		Stream:     stream,
		writeQueue: make(chan []byte, conf.C.ConnWriteQueueSize),
		processor:  processor,
		logicQueue: make(chan []byte, conf.C.ConnUndoQueueSize),
	}
	// write q
	go func() {
		for pkg := range s.writeQueue {
			if pkg == nil {
				break
			}
			if conf.C.ConnWriteTimeout > 0 {
				_ = s.SetWriteDeadline(time.Now().Add(time.Second * time.Duration(conf.C.ConnWriteTimeout)))
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

func (tc *QuicSteam) GetUuid() string {
	return tc.uuid
}

// read | write end -> write | read end -> conn end
func (s *QuicSteam) ReadMsg() {
	defer func() {
		s.logicQueue <- nil
		s.writeQueue <- nil
		// force close conn
		if !s.IsClosed() {
			_ = s.Close()
		}
	}()
	bf := make([]byte, conf.C.MaxDataPackageSize)
	// 第一个包默认5秒
	timeout := time.Second * time.Duration(conf.C.FirstPackageTimeout)
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
		if ln < 1 || int(ln) > conf.C.MaxDataPackageSize {
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
		timeout = time.Second * time.Duration(conf.C.ConnReadTimeout)
	}
}

func (s *QuicSteam) WriteMsg(message interface{}) {
	err, pkg := s.processor.WarpMsg(message)
	if err != nil {
		log.Error("Quic Steam OnWarpMsg package error %s", err)
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

func (s *QuicSteam) Close() error {
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

func (s *QuicSteam) IsClosed() bool {
	return atomic.LoadInt64(&s.closeFlag) != 0
}

func (s *QuicSteam) AfterClose(cb func()) {
	s.Lock()
	defer s.Unlock()
	s.closeCb = cb
}

func (s *QuicSteam) SetData(data interface{}) {
	s.Lock()
	defer s.Unlock()
	s.userData = data
}
func (s *QuicSteam) GetData() interface{} {
	s.RLock()
	defer s.RUnlock()
	return s.userData
}
func (s *QuicSteam) SetNode(node iduck.INode) {
	s.Lock()
	defer s.Unlock()
	s.node = node
}
func (s *QuicSteam) GetNode() iduck.INode {
	s.RLock()
	defer s.RUnlock()
	return s.node
}
