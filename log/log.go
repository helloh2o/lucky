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
	warnLevel    = 2
	errorLevel   = 3
	fatalLevel   = 4
)

const (
	printDebugLevel   = "[debug  ] "
	printReleaseLevel = "[release] "
	printWarnLevel    = "[warn   ] "
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
		level = warnLevel
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
		newLogLevel = warnLevel
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

func (logger *Logger) doPrintf(level int, printLevel string, a ...interface{}) {
	empty := ""
	if level < logger.GetOutputLv() || len(a) == 0 {
		return
	}
	format := empty
	if len(a) > 1 {
		format, _ = a[0].(string)
	}
	if logger.baseLogger == nil {
		panic("logger closed")
	}
	content := empty
	if format != empty {
		format = printLevel + format
		content = fmt.Sprintf(format, a[1:]...)
	} else {
		sb := strings.Builder{}
		for _, v := range a {
			sb.WriteString(fmt.Sprintf("%+v", v))
		}
		content = printLevel + sb.String()
	}
	logger.baseLogger.Output(3, content)

	if level == fatalLevel {
		os.Exit(1)
	}
}

// Debug log
func (logger *Logger) Debug(a ...interface{}) {
	logger.doPrintf(debugLevel, printDebugLevel, a...)
}

// Release log
func (logger *Logger) Release(a ...interface{}) {
	logger.doPrintf(releaseLevel, printReleaseLevel, a...)
}

// Warn log
func (logger *Logger) Warn(a ...interface{}) {
	logger.doPrintf(warnLevel, printWarnLevel, a...)
}

// Error log
func (logger *Logger) Error(a ...interface{}) {
	logger.doPrintf(errorLevel, printErrorLevel, a...)
}

// Fatal panic
func (logger *Logger) Fatal(a ...interface{}) {
	logger.doPrintf(fatalLevel, printFatalLevel, a...)
}

// Export It's dangerous to call the method on logging
func Export(logger *Logger) {
	if logger != nil {
		defaultLogger = logger
	}
}

// Debug print
func Debug(a ...interface{}) {
	defaultLogger.doPrintf(debugLevel, printDebugLevel, a...)
}

// Release print
func Release(a ...interface{}) {
	defaultLogger.doPrintf(releaseLevel, printReleaseLevel, a...)
}

// Warn print
func Warn(a ...interface{}) {
	defaultLogger.doPrintf(warnLevel, printWarnLevel, a...)
}

// Error print
func Error(a ...interface{}) {
	defaultLogger.doPrintf(errorLevel, printErrorLevel, a...)
}

// Fatal print
func Fatal(a ...interface{}) {
	defaultLogger.doPrintf(fatalLevel, printFatalLevel, a...)
}

// Close default logger
func Close() {
	defaultLogger.Close()
}
