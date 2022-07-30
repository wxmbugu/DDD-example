package logger

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
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
	Out      io.Writer
	Level    Level
	Mutex    sync.Mutex
	ExitFunc exitFunc
}

type exitFunc func(int)

func New() *Logger {
	return &Logger{
		Out:   os.Stderr,
		Level: LevelInfo,
	}
}

func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelDebug:
		return "DEBUG"
	case LevelWarning:
		return "WARNING"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "TRACE"
	}
}

var errprint = color.New(color.FgRed).Add(color.Bold).PrintFunc()
var infoprint = color.New(color.FgYellow).Add(color.Bold).PrintFunc()
var fatalprint = color.New(color.FgRed).Add(color.Bold).PrintFunc()

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
	if level < l.Level {
		return 0, nil
	}
	data := struct {
		Level   string
		Time    string
		Message string
		Data    interface{}
	}{
		Level:   level.String(),
		Time:    time.Now().UTC().Format(time.RFC3339),
		Message: message,
		Data:    properties,
	}
	var line []byte
	line, err := json.Marshal(&data)
	if err != nil {
		line = []byte(LevelError.String() + ": unable to marshal log message" + err.Error())
	}
	l.Mutex.Lock()
	defer l.Mutex.Unlock()
	return l.Out.Write(append(line, '\n'))
}

func (l *Logger) Write(message []byte) (int, error) {
	return l.print(LevelError, string(message), nil)
}
