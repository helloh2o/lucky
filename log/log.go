package log

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

// levels
const (
	debugLevel   = 0
	releaseLevel = 1
	errorLevel   = 2
	fatalLevel   = 3
)

const (
	printDebugLevel   = "[debug  ] "
	printReleaseLevel = "[release] "
	printErrorLevel   = "[error  ] "
	printFatalLevel   = "[fatal  ] "
)

type Logger struct {
	level      int
	baseLogger *log.Logger
	baseFile   *os.File
}

var defaultLogger *Logger

func init() {
	// new
	defaultLogger = &Logger{
		level:      debugLevel,
		baseLogger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}
func New(strLevel string, pathname string, flag int) (*Logger, error) {
	// level
	var level int
	switch strings.ToLower(strLevel) {
	case "debug":
		level = debugLevel
	case "release":
		level = releaseLevel
	case "error":
		level = errorLevel
	case "fatal":
		level = fatalLevel
	default:
		return nil, errors.New("unknown level: " + strLevel)
	}

	// logger
	var baseLogger *log.Logger
	var baseFile *os.File
	if pathname != "" {
		now := time.Now()

		filename := fmt.Sprintf("%d%02d%02d_%02d_%02d_%02d.log",
			now.Year(),
			now.Month(),
			now.Day(),
			now.Hour(),
			now.Minute(),
			now.Second())

		file, err := os.Create(path.Join(pathname, filename))
		if err != nil {
			return nil, err
		}

		baseLogger = log.New(io.MultiWriter(file, os.Stdout), "", flag)
		baseFile = file
	} else {
		baseLogger = log.New(os.Stdout, "", flag)
	}
	// new
	logger := new(Logger)
	logger.level = level
	logger.baseLogger = baseLogger
	logger.baseFile = baseFile
	// replace default logger
	defaultLogger = logger
	return logger, nil
}

// It's dangerous to call the method on logging
func (logger *Logger) Close() {
	if logger.baseFile != nil {
		logger.baseFile.Close()
	}

	logger.baseLogger = nil
	logger.baseFile = nil
}

func (logger *Logger) doPrintf(level int, printLevel string, format string, a ...interface{}) {
	if level < logger.level {
		return
	}
	if logger.baseLogger == nil {
		panic("logger closed")
	}

	format = printLevel + format
	logger.baseLogger.Output(3, fmt.Sprintf(format, a...))

	if level == fatalLevel {
		os.Exit(1)
	}
}

func (logger *Logger) Debug(format string, a ...interface{}) {
	logger.doPrintf(debugLevel, printDebugLevel, format, a...)
}

func (logger *Logger) Release(format string, a ...interface{}) {
	logger.doPrintf(releaseLevel, printReleaseLevel, format, a...)
}

func (logger *Logger) Error(format string, a ...interface{}) {
	logger.doPrintf(errorLevel, printErrorLevel, format, a...)
}

func (logger *Logger) Fatal(format string, a ...interface{}) {
	logger.doPrintf(fatalLevel, printFatalLevel, format, a...)
}

// It's dangerous to call the method on logging
func Export(logger *Logger) {
	if logger != nil {
		defaultLogger = logger
	}
}

func Debug(format string, a ...interface{}) {
	defaultLogger.doPrintf(debugLevel, printDebugLevel, format, a...)
}

func Release(format string, a ...interface{}) {
	defaultLogger.doPrintf(releaseLevel, printReleaseLevel, format, a...)
}

func Error(format string, a ...interface{}) {
	defaultLogger.doPrintf(errorLevel, printErrorLevel, format, a...)
}

func Fatal(format string, a ...interface{}) {
	defaultLogger.doPrintf(fatalLevel, printFatalLevel, format, a...)
}

func Close() {
	defaultLogger.Close()
}
