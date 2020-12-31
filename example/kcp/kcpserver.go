package main

import (
	"github.com/helloh2o/lucky/core/inet"
	"github.com/helloh2o/lucky/example/comm/msg"
	"log"
)

func main() {
	msg.SetEncrypt(msg.Processor)
	if s, err := inet.NewKcpServer("localhost:2023", msg.Processor); err != nil {
		panic(err)
	} else {
		log.Fatal(s.Run())
	}
}
