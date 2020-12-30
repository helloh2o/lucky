package inet

import (
	"github.com/helloh2o/lucky/core/iduck"
	"github.com/helloh2o/lucky/log"
	"net"
	"runtime/debug"
	"sync"
)

type tcpServer struct {
	mu        sync.Mutex
	addr      string
	ln        net.Listener
	processor iduck.Processor
}

func NewTcpServer(addr string, processor iduck.Processor) (s *tcpServer, err error) {
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

// goroutine handle connection
func (s *tcpServer) Handle(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("PANIC %v TCP handle, stack %s", r, string(debug.Stack()))
		}
	}()
	var ic iduck.IConnection
	ic = NewTcpConn(conn, s.processor)
	ic.ReadMsg()
}
