package logging

import (
	"log"
	"os"
	"runtime/debug"
)

type LoggerSet struct {
	Debug   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
	Fatal   *log.Logger
}

var (
	loggerSet LoggerSet
)

func init() {
	loggerSet.Debug = log.New(os.Stdout, "[DEBUG]   ", log.Ldate|log.Lmicroseconds|log.Lmsgprefix)
	loggerSet.Info = log.New(os.Stdout, "[INFO]    ", log.Ldate|log.Lmicroseconds|log.Lmsgprefix)
	loggerSet.Warning = log.New(os.Stdout, "[WARNING] ", log.Ldate|log.Lmicroseconds|log.Lmsgprefix)
	loggerSet.Error = log.New(os.Stderr, "[ERROR]   ", log.Ldate|log.Lmicroseconds|log.Lmsgprefix)
	loggerSet.Fatal = log.New(os.Stderr, "[FATAL]   ", log.Ldate|log.Lmicroseconds|log.Lmsgprefix)
}

func Debug(format string, args ...any) {
	loggerSet.Debug.Printf(format, args...)
}

func Info(format string, args ...any) {
	loggerSet.Info.Printf(format, args...)
}

func Warning(format string, args ...any) {
	loggerSet.Warning.Printf(format, args...)
}

func Error(format string, args ...any) {
	loggerSet.Error.Printf(format, args...)
}

func Fatal(format string, args ...any) {
	loggerSet.Error.Printf(format, args...)
	loggerSet.Error.Fatal(string(debug.Stack()))
}
