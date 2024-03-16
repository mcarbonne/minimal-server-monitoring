package logging

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
)

var (
	defaultLogger = log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds|log.Lmsgprefix)
)

func Log(level, format string, args ...any) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	} else {
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				file = file[i+1:]
				break
			}
		}
	}
	defaultLogger.Printf("%-6v %v:%v\t%v", level, file, line, fmt.Sprintf(format, args...))
}

func Debug(format string, args ...any) {
	Log("DEBUG", format, args...)
}

func Info(format string, args ...any) {
	Log("INFO", format, args...)
}

func Warning(format string, args ...any) {
	Log("WARN", format, args...)
}

func Error(format string, args ...any) {
	Log("ERROR", format, args...)
}

func Fatal(format string, args ...any) {
	Log("FATAL", format, args...)
	defaultLogger.Fatal(string(debug.Stack()))
}
