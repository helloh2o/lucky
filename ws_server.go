package lucky

import (
	"github.com/gorilla/websocket"
	"github.com/helloh2o/lucky/log"
	"net"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

type wsServer struct {
	mu        sync.Mutex
	addr      string
	ln        net.Listener
	processor Processor
}

type wsHandler struct {
	sv       *wsServer
	upgrader websocket.Upgrader
}

func (h *wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("upgrade error: %v", err)
		return
	}
	go h.HandlerWsConn(conn)
}

// HandlerWsConn goroutine handle connection
func (h *wsHandler) HandlerWsConn(conn *websocket.Conn) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("PANIC %v websocket handle, stack %s", r, string(debug.Stack()))
		}
	}()
	var ic IConnection
	ic = NewWSConn(conn, h.sv.processor)
	ic.ReadMsg()
}

// NewWsServer return new wsServer
func NewWsServer(addr string, processor Processor) (s *wsServer, err error) {
	wss := new(wsServer)
	wss.addr = addr
	wss.ln, err = net.Listen("tcp", addr)
	if processor == nil {
		panic("processor must be set.")
	}
	wss.processor = processor
	if err != nil {
		return nil, err
	}
	return wss, err
}

// Run the ws server
func (s *wsServer) Run() error {
	log.Release("Starting websocket server on %s", s.addr)
	httpServer := &http.Server{
		Addr: s.addr,
		Handler: &wsHandler{sv: s, upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}},
		ReadTimeout:    time.Second * time.Duration(C.ConnReadTimeout),
		WriteTimeout:   time.Second * time.Duration(C.ConnWriteTimeout),
		MaxHeaderBytes: C.MaxHeaderLen,
	}
	return httpServer.Serve(s.ln)
}

// Handle
func (s *wsServer) Handle(conn net.Conn) {}
