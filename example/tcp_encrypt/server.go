package main

import (
	"github.com/sirupsen/logrus"
	"net/http"

	//stdlog "log"
	"lucky-day/core/inet"
	"lucky-day/example/tcp_encrypt/msg"
	_ "net/http/pprof"
)

func main() {
	go func() {
		//go tool pprof  http://localhost:6060/debug/pprof/profile
		_ = http.ListenAndServe("0.0.0.0:6060", nil)
	}()
	//_, err := log.New("debug", ".", stdlog.LstdFlags|stdlog.Lshortfile)
	if s, err := inet.NewTcpServer("localhost:2021", msg.Processor); err != nil {
		panic(err)
	} else {
		err = s.Run()
		logrus.Print(err)
	}
}
