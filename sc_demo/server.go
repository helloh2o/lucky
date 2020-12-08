package main

import (
	"github.com/sirupsen/logrus"
	//stdlog "log"
	"lucky-day/core/inet"
	"lucky-day/sc_demo/msg"
)

func main() {
	//_, err := log.New("debug", ".", stdlog.LstdFlags|stdlog.Lshortfile)
	if s, err := inet.NewTcpServer("localhost:2021", msg.Processor); err != nil {
		panic(err)
	} else {
		err = s.Run()
		logrus.Print(err)
	}
}
