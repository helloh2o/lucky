package inet

import (
	"github.com/xtaci/kcp-go"
	"lucky/core/iduck"
	"lucky/log"
	"net"
	"runtime/debug"
	"sync"
)

type kcpServer struct {
	mu        sync.Mutex
	addr      string
	ln        net.Listener
	processor iduck.Processor
}

func NewKcpServer(addr string, processor iduck.Processor) (s *tcpServer, err error) {
	ts := new(tcpServer)
	ts.addr = addr
	ts.ln, err = kcp.ListenWithOptions(":2023", nil, 10, 3)
	if processor == nil {
		panic("processor must be set.")
	}
	ts.processor = processor
	if err != nil {
		return nil, err
	}
	return ts, err
}

func (s *kcpServer) Run() error {
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
func (s *kcpServer) Handle(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("PANIC %v TCP handle, stack %s", r, string(debug.Stack()))
		}
	}()
	var ic iduck.IConnection
	ic = NewTcpConn(conn, s.processor)
	ic.ReadMsg()
}
