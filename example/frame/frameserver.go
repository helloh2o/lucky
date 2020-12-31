package main

import (
	"github.com/helloh2o/lucky/core/inet"
	"github.com/helloh2o/lucky/example/comm/msg"
	"github.com/helloh2o/lucky/example/comm/node"
	"log"
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
		log.Fatal(s.Run())
	}
}
