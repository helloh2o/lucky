package main

import (
	"lucky/core/inet"
	"lucky/example/comm/msg"
	"lucky/example/comm/node"
)

func main() {
	msg.SetEncrypt(msg.Processor)
	node.NewTestNode()
	if s, err := inet.NewKcpServer("localhost:2024", msg.Processor); err != nil {
		panic(err)
	} else {
		err = s.Run()
	}
}
