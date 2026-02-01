package logging

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
)

type ErrorHook struct {
	fn    func(string)
	mutex sync.RWMutex
}

var (
	defaultLogger = log.New(os.Stdout, "", log.Ldate|log.Lmicroseconds|log.Lmsgprefix)
	errorHook     ErrorHook
)

func SetErrorHook(hook func(string)) {
	errorHook.mutex.Lock()
	defer errorHook.mutex.Unlock()
	errorHook.fn = hook
}

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
	msg := fmt.Sprintf(format, args...)
	Log("ERROR", "%s", msg)
	errorHook.mutex.RLock()
	defer errorHook.mutex.RUnlock()
	if errorHook.fn != nil {
		errorHook.fn(msg)
	}
}

func Fatal(format string, args ...any) {
	Log("FATAL", format, args...)
	defaultLogger.Fatal(string(debug.Stack()))
}
