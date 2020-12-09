package main

import (
	"lucky/log"
	"net/http"

	stdlog "log"
	"lucky/core/inet"
	"lucky/example/comm/msg"
	_ "net/http/pprof"
)

func main() {
	go func() {
		//go tool pprof  http://localhost:6060/debug/pprof/profile
		_ = http.ListenAndServe("0.0.0.0:6060", nil)
	}()
	_, err := log.New("release", ".", stdlog.LstdFlags|stdlog.Lshortfile)
	if err != nil {
		panic(err)
	}
	msg.SetEncrypt(msg.Processor)
	if s, err := inet.NewTcpServer("localhost:2021", msg.Processor); err != nil {
		panic(err)
	} else {
		err = s.Run()
	}
}
