package logger

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Level uint8

// PERF: pkg/Logger
// TEST: pkg/logger
// TODO: Writeto<write to json file or console or log file> functionality making it extensible depending on developer
const (
	LevelInfo Level = iota
	LevelWarning
	LevelDebug
	LevelFatal
	LevelError
	LevelTrace
)

type Logger struct {
	out   io.Writer
	Level Level
	Mutex sync.Mutex
	Color bool
}

func New() *Logger {
	return &Logger{
		out:   os.Stderr,
		Level: LevelInfo,
	}
}

func (l Level) stringformat() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelDebug:
		return "DEBG"
	case LevelWarning:
		return "WARN"
	case LevelError:
		return "EROR"
	case LevelFatal:
		return "FATL"
	default:
		return "TRAC"
	}
}

// "\033[1;31m"
const (
	Reset = iota
	Red   = iota + 30
	Green
	Yellow
	Blue
	Purple
	Cyan
)

var colors map[Level]string

func colorformat(level Level, color int) string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s[%v]", reset(), level.stringformat())
	}
	return fmt.Sprintf("\033[%dm[%v]", color, level.stringformat())
}
func reset() string {
	return fmt.Sprintf("\033[%dm\n", Reset)
}

func (l Level) colored() string {
	switch l {
	case LevelInfo:
		return colorformat(LevelInfo, Cyan)
	case LevelDebug:
		return colorformat(LevelDebug, Yellow)
	case LevelWarning:
		return colorformat(LevelWarning, Purple)
	case LevelError:
		return colorformat(LevelError, Red)
	case LevelFatal:
		return colorformat(LevelFatal, Red)
	case LevelTrace:
		return colorformat(LevelTrace, Blue)
	default:
		return ""
	}
}

func (l *Logger) Info(message string, properties ...interface{}) {
	l.print(LevelInfo, message, properties)
}
func (l *Logger) Error(err error, properties ...interface{}) {
	l.print(LevelError, err.Error(), properties)
}
func (l *Logger) Fatal(err error, properties ...interface{}) {
	l.print(LevelFatal, err.Error(), properties)
	os.Exit(1) // For entries at the FATAL level, we also terminate the application.
}
func (l *Logger) Debug(message string, properties ...interface{}) {
	l.print(LevelDebug, message, properties)
}
func (l *Logger) Warning(err error, properties ...interface{}) {
	l.print(LevelWarning, err.Error(), properties)
}
func (l *Logger) Trace(err error, properties ...interface{}) {
	l.print(LevelTrace, err.Error(), properties)
	os.Exit(1)
}

func (l *Logger) print(level Level, message string, properties ...interface{}) (int, error) {
	if level < l.Level {
		return 0, nil
	}
	var buf []byte
	buf = append(buf, level.colored()...)
	buf = append(buf, " "...)
	buf = append(buf, time.Now().Format("2006-01-02T15:04:05Z")...)
	buf = append(buf, " "...)
	buf = append(buf, message...)
	buf = append(buf, " "...)
	props := fmt.Sprintf("%v", properties)
	buf = append(buf, strings.Trim(props, "[]")...)
	buf = append(buf, reset()...)
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	return l.out.Write(buf)
}

func (l *Logger) Write(message []byte) (int, error) {
	return l.print(LevelError, string(message), nil)
}
