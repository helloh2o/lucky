package inet

import (
	"github.com/gorilla/websocket"
	"lucky-day/core/iduck"
	"lucky-day/log"
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
	processor iduck.Processor
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

// goroutine handle connection
func (h *wsHandler) HandlerWsConn(conn *websocket.Conn) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("PANIC %v TCP handle, stack %s", r, string(debug.Stack()))
		}
	}()
	var ic iduck.IConnection
	ic = NewWSConn(conn, h.sv.processor, 100)
	ic.ReadMsg()
}

func NewWsServer(addr string, processor iduck.Processor) (s *wsServer, err error) {
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

func (s *wsServer) Run() error {
	log.Release("Starting websocket server on %s", s.addr)
	httpServer := &http.Server{
		Addr:           s.addr,
		Handler:        &wsHandler{sv: s},
		ReadTimeout:    time.Second * 10,
		WriteTimeout:   time.Second * 10,
		MaxHeaderBytes: 1024,
	}
	return httpServer.Serve(s.ln)
}

func (s *wsServer) Handle(conn net.Conn) {}
