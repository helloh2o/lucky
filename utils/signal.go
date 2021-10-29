package utils

import (
	"github.com/helloh2o/lucky/log"
	"os"
	"os/signal"
	"syscall"
)

func SignalExit(callback func()) {
	// 监听信号
	exit := make(chan os.Signal)
	signal.Notify(exit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
	for {
		s := <-exit
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT:
			log.Release("Exit %s", s)
			if callback != nil {
				callback()
			}
			break
		default:
			log.Error("Got signal %s, do you want to exit? ", s)
		}
	}
}
