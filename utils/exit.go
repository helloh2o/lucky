package utils

import (
	"context"
	"github.com/helloh2o/lucky"
	"github.com/helloh2o/lucky/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// GraceExit 优雅退出
func GraceExit(done chan struct{}, callbacks ...func()) {
	//等待信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGKILL)
	log.Release("wait exit signal before graceful shutdown ...")
	s := <-quit
	log.Release("process %d exit gracefully by signal %s", os.Getpid(), s.String())
	for _, f := range callbacks {
		f()
	}
	done <- struct{}{}
}

// IrisSVExit Iris服务优雅退出
func IrisSVExit(done chan struct{}, callbacks ...func()) {
	callbacks = append(callbacks, func() {
		timeout := 10 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		// 关闭所有主机
		_ = lucky.Iris().Shutdown(ctx)
		log.Release("iris is shutdown gracefully")
	}, func() {
		os.Exit(0)
	})
	GraceExit(done, callbacks...)
}
