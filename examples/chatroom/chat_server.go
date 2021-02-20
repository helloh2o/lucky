package main

import (
	"github.com/helloh2o/lucky"
	"github.com/helloh2o/lucky/examples/chatroom/jsonmsg"
	"log"
	"net/http"

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
		FirstPackageTimeout: 5,
		ConnReadTimeout:     15,
		ConnWriteTimeout:    5,
		MaxDataPackageSize:  2048,
		MaxHeaderLen:        1024,
	})
	if s, err := lucky.NewWsServer("localhost:20220", jsonmsg.Processor); err != nil {
		panic(err)
	} else {
		log.Fatal(s.Run())
	}
}
