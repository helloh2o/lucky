package lucky

import (
	"context"
	"crypto/tls"
	"github.com/helloh2o/lucky/log"
	"github.com/lucas-clemente/quic-go"
	"net"
	"runtime/debug"
	"sync"
)

type quicServer struct {
	mu        sync.Mutex
	addr      string
	ln        quic.Listener
	processor Processor
}

// NewQUICServer return new quicServer
func NewQUICServer(addr string, processor Processor, config *tls.Config) (s *quicServer, err error) {
	ts := new(quicServer)
	ts.addr = addr
	ts.ln, err = quic.ListenAddr(addr, config, nil)
	if processor == nil {
		panic("processor must be set.")
	}
	ts.processor = processor
	if err != nil {
		return nil, err
	}
	return ts, err
}

func (s *quicServer) Run() error {
	log.Release("Starting quic server on %s", s.addr)
	for {
		sess, err := s.ln.Accept(context.Background())
		if err != nil {
			return err
		}
		stream, err := sess.AcceptStream(context.Background())
		if err != nil {
			log.Error("Accept stream from session error %v", err)
			continue
		}
		go s.HandleStream(stream)
	}
}

// goroutine handle Stream
func (s *quicServer) HandleStream(stream quic.Stream) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("PANIC %v quic stream handle, stack %s", r, string(debug.Stack()))
		}
	}()
	var ic IConnection
	// 可靠安全的UDP协议，http/3
	ic = NewQuicStream(stream, s.processor)
	ic.ReadMsg()
}
func (s *quicServer) Handle(conn net.Conn) {}
