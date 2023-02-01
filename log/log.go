package log

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"sync"
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

// Logger warp
type Logger struct {
	mu         sync.RWMutex
	level      int
	baseLogger *log.Logger
	BaseFile   *os.File
}

var defaultLogger *Logger

func init() {
	// new
	defaultLogger = &Logger{
		level:      debugLevel,
		baseLogger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
}

// New a logger
func New(strLevel string, pathname string, flag int) (*Logger, error) {
	// level
	var level int
	switch strings.ToLower(strLevel) {
	case "debug":
		level = debugLevel
	case "release":
		level = releaseLevel
	case "warn":
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
	logger.BaseFile = baseFile
	// replace default logger
	defaultLogger = logger
	return logger, nil
}

// Close It's dangerous to call the method on logging
func (logger *Logger) Close() {
	if logger.BaseFile != nil {
		logger.BaseFile.Close()
	}

	logger.baseLogger = nil
	logger.BaseFile = nil
}

// read log lv
func (logger *Logger) GetOutputLv() int {
	logger.mu.RLock()
	lv := logger.level
	defer logger.mu.RUnlock()
	return lv
}

// update log level
func (logger *Logger) SetLogLevel(strLevel string) {
	logger.mu.Lock()
	defer logger.mu.Unlock()
	// newLogLevel
	var newLogLevel int
	switch strings.ToLower(strLevel) {
	case "debug":
		newLogLevel = debugLevel
	case "release":
		newLogLevel = releaseLevel
	case "warn":
		newLogLevel = releaseLevel
	case "error":
		newLogLevel = errorLevel
	case "fatal":
		newLogLevel = fatalLevel
	default:
		newLogLevel = logger.level
	}
	logger.level = newLogLevel
}

func SetLogLevelDefault(lv string) {
	defaultLogger.SetLogLevel(lv)
}

func (logger *Logger) doPrintf(level int, printLevel string, format string, a ...interface{}) {
	if level < logger.GetOutputLv() {
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

// Debug log
func (logger *Logger) Debug(format string, a ...interface{}) {
	logger.doPrintf(debugLevel, printDebugLevel, format, a...)
}

// Release log
func (logger *Logger) Release(format string, a ...interface{}) {
	logger.doPrintf(releaseLevel, printReleaseLevel, format, a...)
}

// Error log
func (logger *Logger) Error(format string, a ...interface{}) {
	logger.doPrintf(errorLevel, printErrorLevel, format, a...)
}

// Fatal panic
func (logger *Logger) Fatal(format string, a ...interface{}) {
	logger.doPrintf(fatalLevel, printFatalLevel, format, a...)
}

// Export It's dangerous to call the method on logging
func Export(logger *Logger) {
	if logger != nil {
		defaultLogger = logger
	}
}

// Debug print
func Debug(format string, a ...interface{}) {
	defaultLogger.doPrintf(debugLevel, printDebugLevel, format, a...)
}

// Release print
func Release(format string, a ...interface{}) {
	defaultLogger.doPrintf(releaseLevel, printReleaseLevel, format, a...)
}

// Error print
func Error(format string, a ...interface{}) {
	defaultLogger.doPrintf(errorLevel, printErrorLevel, format, a...)
}

// Fatal print
func Fatal(format string, a ...interface{}) {
	defaultLogger.doPrintf(fatalLevel, printFatalLevel, format, a...)
}

// Close default logger
func Close() {
	defaultLogger.Close()
}
