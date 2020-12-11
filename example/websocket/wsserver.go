package main

import (
	"lucky/conf"
	"net/http"

	"lucky/core/inet"
	"lucky/example/comm/msg"
	_ "net/http/pprof"
)

func main() {
	go func() {
		//go tool pprof -http=:1234 http://localhost:6060/debug/pprof/profile
		_ = http.ListenAndServe("0.0.0.0:6060", nil)
	}()
	/*_, err := log.New("release", ".", stdlog.LstdFlags|stdlog.Lshortfile)
	if err != nil {
		panic(err)
	}*/
	conf.Set(&conf.Data{
		ConnUndoQueueSize:   100,
		ConnWriteQueueSize:  100,
		FirstPackageTimeout: 5,
		ConnReadTimeout:     15,
		ConnWriteTimeout:    5,
		MaxDataPackageSize:  2048,
		MaxHeaderLen:        1024,
	})
	msg.SetEncrypt(msg.Processor)
	if s, err := inet.NewWsServer("localhost:2022", msg.Processor); err != nil {
		panic(err)
	} else {
		err = s.Run()
	}
}
