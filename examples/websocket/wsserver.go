package main

import (
	"github.com/helloh2o/lucky"
	"log"
	"net/http"

	"github.com/helloh2o/lucky/examples/comm/msg"
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
	lucky.SetConf(&lucky.Data{
		ConnUndoQueueSize:   100,
		ConnWriteQueueSize:  100,
		FirstPackageTimeout: 500,
		ConnReadTimeout:     500,
		ConnWriteTimeout:    5,
		MaxDataPackageSize:  2048,
		MaxHeaderLen:        1024,
	})
	msg.SetEncrypt(msg.Processor)
	if s, err := lucky.NewWsServer(":2022", msg.Processor); err != nil {
		panic(err)
	} else {
		log.Fatal(s.Run())
	}
}
