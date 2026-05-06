package logafa

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/fatih/color"
)
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

var CurrentLevel LogLevel = DEBUG
var LogFile *os.File

func Debug(format string, args ...interface{}) { logMessage(DEBUG, "DEBUG", fmt.Sprintf(format, args...)) }
func Info(format string, args ...interface{})  { logMessage(INFO, "INFO", fmt.Sprintf(format, args...)) }
func Warn(format string, args ...interface{})  { logMessage(WARN, "WARN", fmt.Sprintf(format, args...)) }
func Error(format string, args ...interface{}) { logMessage(ERROR, "ERROR", fmt.Sprintf(format, args...)) }


func logMessage(level LogLevel, levelStr, msg string) {
	if level < CurrentLevel {
		return
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	
	// 取得呼叫者資訊
	_, file, line, ok := runtime.Caller(2)
	location := ""
	if ok {
		location = fmt.Sprintf("%s:%d", file, line)
	}

	prefix := fmt.Sprintf("[%s] [%s] [%s]", timestamp, levelStr, location)

	var coloredMsg string
	switch level {
	case DEBUG:
		coloredMsg = color.New(color.FgCyan).Sprint(msg)
	case INFO:
		coloredMsg = color.New(color.FgGreen).Sprint(msg)
	case WARN:
		coloredMsg = color.New(color.FgYellow).Sprint(msg)
	case ERROR:
		coloredMsg = color.New(color.FgRed).Sprint(msg)
	default:
		coloredMsg = msg
	}

	formatted := fmt.Sprintf("%s %s\n", prefix, coloredMsg)
	fmt.Print(formatted)
	// 寫入檔案
	if LogFile != nil {
		LogFile.WriteString(formatted)
	}
}