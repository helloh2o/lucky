package main

import (
	"github.com/helloh2o/lucky"
	"github.com/helloh2o/lucky/examples/chatroom/jsonmsg"
	"github.com/helloh2o/lucky/log"
)

func main() {
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
		log.Fatal("%v", s.Run())
	}
}
