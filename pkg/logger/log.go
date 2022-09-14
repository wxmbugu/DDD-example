package logger

import (
	//"encoding/json"
	//	"bytes"
	//"encoding/gob"
	//"bytes"
	//"encoding/gob"
	"fmt"
	//"image/color"
	"io"
	"os"
	"sync"
	"time"
)

type Level uint8

const (
	LevelInfo Level = iota
	LevelWarning
	LevelDebug
	LevelFatal
	LevelError
	LevelTrace
)

type Logger struct {
	Out   io.Writer
	Level Level
	Mutex sync.Mutex
	Color bool
}

func New() *Logger {
	return &Logger{
		Out:   os.Stderr,
		Level: LevelInfo,
	}
}

func (l Level) stringformat() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelDebug:
		return "DEBUG"
	case LevelWarning:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "TRACE"
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
	return fmt.Sprintf("\033[%dm%v", color, level.stringformat())
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

func (l *Logger) PrintInfo(message string, properties ...interface{}) {
	l.print(LevelInfo, message, properties)
}
func (l *Logger) PrintError(err error, properties ...interface{}) {
	l.print(LevelError, err.Error(), properties)
}
func (l *Logger) PrintFatal(err error, properties ...interface{}) {
	l.print(LevelFatal, err.Error(), properties)
	os.Exit(1) // For entries at the FATAL level, we also terminate the application.
}
func (l *Logger) PrintDebug(message string, properties ...interface{}) {
	l.print(LevelDebug, message, properties)
}
func (l *Logger) PrintWarning(err error, properties ...interface{}) {
	l.print(LevelWarning, err.Error(), properties)
}
func (l *Logger) PrintTrace(err error, properties ...interface{}) {
	l.print(LevelTrace, err.Error(), properties)
	os.Exit(1) // For entries at the FATAL level, we also terminate the application.
}

func (l *Logger) print(level Level, message string, properties ...interface{}) (int, error) {
	//fmt.Println(level.String())
	//	fmt.Println(colors[LevelInfo])
	if level < l.Level {
		return 0, nil
	}

	// _ = struct {
	// 	Level   string
	// 	Time    string
	// 	Message string
	// 	Data    interface{}
	// }{
	// 	Level:   level.String(),
	// 	Time:    time.Now().UTC().Format(time.RFC1123),
	// 	Message: message,
	// 	Data:    properties,
	// }

	var buf []byte
	buf = append(buf, level.colored()...)
	buf = append(buf, " "...)
	buf = append(buf, time.Now().UTC().Format(time.RFC1123)...)
	buf = append(buf, " "...)
	buf = append(buf, message...)
	buf = append(buf, " "...)
	buf = append(buf, reset()...)

	// var buf1 bytes.Buffer
	// enc := gob.NewEncoder(&buf1)
	// _ = enc.Encode(properties)
	// data := buf1.Bytes()
	// buf = append(buf, data...)
	l.Mutex.Lock()

	defer l.Mutex.Unlock()
	return l.Out.Write(buf)
}

func (l *Logger) Write(message []byte) (int, error) {
	return l.print(LevelError, string(message), nil)
}
