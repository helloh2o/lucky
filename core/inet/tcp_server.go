package inet

import (
	"github.com/sirupsen/logrus"
	"lucky-day/core/duck"
	"net"
	"runtime/debug"
	"sync"
)

type tcpServer struct {
	mu        sync.Mutex
	addr      string
	ln        net.Listener
	Conns     map[interface{}]duck.IConnection
	processor duck.Processor
}

func NewTcpServer(addr string, processor duck.Processor) (s *tcpServer, err error) {
	ts := new(tcpServer)
	ts.addr = addr
	ts.ln, err = net.Listen("tcp", addr)
	if processor == nil {
		panic("processor must be set.")
	}
	ts.processor = processor
	ts.Conns = make(map[interface{}]duck.IConnection)
	if err != nil {
		return nil, err
	}
	return ts, err
}

func (s *tcpServer) Run() error {
	logrus.Infof("Starting tcp server on %s", s.addr)
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return err
		}
		go s.Handle(conn)
	}
}

// goroutine handle connection
func (s *tcpServer) Handle(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("PANIC %v TCP handle, stack %s", r, string(debug.Stack()))
		}
		s.mu.Lock()
		delete(s.Conns, conn.RemoteAddr())
		s.mu.Unlock()
	}()
	var ic duck.IConnection
	ic = NewTcpConn(conn, s.processor)
	s.mu.Lock()
	s.Conns[conn.RemoteAddr()] = ic
	s.mu.Unlock()
	ic.ReadMsg()
}
