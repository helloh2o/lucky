package lucky

import (
	"github.com/helloh2o/lucky/log"
	"net"
	"runtime/debug"
	"sync"
)

type tcpServer struct {
	mu        sync.Mutex
	addr      string
	ln        net.Listener
	processor Processor
}

// NewTcpServer return new tcpServer
func NewTcpServer(addr string, processor Processor) (s *tcpServer, err error) {
	ts := new(tcpServer)
	ts.addr = addr
	ts.ln, err = net.Listen("tcp", addr)
	if processor == nil {
		panic("processor must be set.")
	}
	ts.processor = processor
	if err != nil {
		return nil, err
	}
	return ts, err
}

// Run the server
func (s *tcpServer) Run() error {
	log.Release("Starting tcp server on %s", s.addr)
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return err
		}
		go s.Handle(conn)
	}
}

// Handle goroutine handle connection
func (s *tcpServer) Handle(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("PANIC %v TCP handle, stack %s", r, string(debug.Stack()))
		}
	}()
	var ic IConnection
	ic = NewTcpConn(conn, s.processor)
	ic.ReadMsg()
}
