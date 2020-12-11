package main

import (
	"lucky/core/inet"
	"lucky/example/comm/msg"
	"lucky/example/comm/node"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	go func() {
		//go tool pprof -http=:1234 http://localhost:6060/debug/pprof/profile
		_ = http.ListenAndServe("0.0.0.0:6060", nil)
	}()
	msg.SetEncrypt(msg.Processor)
	node.NewTestNode()
	if s, err := inet.NewKcpServer("localhost:2024", msg.Processor); err != nil {
		panic(err)
	} else {
		err = s.Run()
	}
}
