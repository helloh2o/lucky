package initialize

import (
	"github.com/helloh2o/lucky"
	"github.com/helloh2o/lucky/im/server/config"
	"github.com/helloh2o/lucky/log"
	"io"
	stdlog "log"
	"os"
)

func InitLog() *os.File {
	cfg := config.Get()
	// 初始化日志
	logger, err := log.New(cfg.LogLevel, cfg.LogFile, stdlog.LstdFlags|stdlog.Lshortfile)
	if err != nil {
		panic(err)
	}
	var output *os.File
	if logger.BaseFile != nil {
		ows := io.MultiWriter(logger.BaseFile, os.Stdout)
		lucky.SetLogOutput(ows)
		output = logger.BaseFile
	} else {
		lucky.SetLogOutput(os.Stdout)
		output = os.Stdout
	}
	log.Release("logger is loaded ... ")
	return output
}
