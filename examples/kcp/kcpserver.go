package main

import (
	"github.com/helloh2o/lucky"
	"github.com/helloh2o/lucky/examples/comm/msg"
	"log"
)

func main() {
	msg.SetEncrypt(msg.Processor)
	if s, err := lucky.NewKcpServer("localhost:2023", msg.Processor); err != nil {
		panic(err)
	} else {
		log.Fatal(s.Run())
	}
}
