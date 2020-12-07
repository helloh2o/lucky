package main

import (
	"github.com/sirupsen/logrus"
	"lucky-day/core/duck"
	"lucky-day/core/inet"
	"lucky-day/core/iproto"
	"lucky-day/servers/test_c_s/protobuf_test"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	p := iproto.NewPBProcessor()
	p.RegisterHandler(2001, &protobuf_test.Hello{}, func(args ...interface{}) {
		msg := args[0].(*protobuf_test.Hello)
		logrus.Println(msg.Hello)
		conn := args[1].(duck.IConnection)
		conn.WriteMsg(msg)
	})
	if s, err := inet.NewTcpServer("localhost:2021", p); err != nil {
		panic(err)
	} else {
		err = s.Run()
		logrus.Print(err)
	}
}
