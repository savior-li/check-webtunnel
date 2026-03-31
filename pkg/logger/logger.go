package logger

import (
	"fmt"
	"time"
)

var DebugEnabled = false

func SetDebug(enabled bool) {
	DebugEnabled = enabled
}

func IsDebug() bool {
	return DebugEnabled
}

func Debug(format string, args ...interface{}) {
	if DebugEnabled {
		msg := fmt.Sprintf(format, args...)
		fmt.Printf("[DEBUG] [%s] %s\n", time.Now().Format("15:04:05"), msg)
	}
}

func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("[INFO] [%s] %s\n", time.Now().Format("15:04:05"), msg)
}

func Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("[WARN] [%s] %s\n", time.Now().Format("15:04:05"), msg)
}

func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("[ERROR] [%s] %s\n", time.Now().Format("15:04:05"), msg)
}
