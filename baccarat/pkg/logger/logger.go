package logger

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"baccarat/config"

	"github.com/sirupsen/logrus"
)

var (
	Log *logrus.Logger
)

type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// 获取调用者信息
	var caller string
	if entry.HasCaller() {
		fileName := entry.Caller.File
		fileNameParts := strings.Split(fileName, "/")
		caller = fmt.Sprintf("%s:%d", fileNameParts[len(fileNameParts)-1], entry.Caller.Line)
	}

	// 格式化时间
	timestamp := entry.Time.Format("2006-01-02 15:04:05")

	// 对齐日志级别
	level := strings.ToUpper(entry.Level.String())
	alignedLevel := fmt.Sprintf("%-5s", level)

	// 构建日志消息
	logMessage := fmt.Sprintf("[%s] %s %s %s() %s\n", 
		timestamp, 
		alignedLevel, 
		caller, 
		getCallerFuncName(entry.Caller),
		entry.Message,
	)

	return []byte(logMessage), nil
}

// 获取函数名
func getCallerFuncName(caller *runtime.Frame) string {
	if caller == nil {
		return ""
	}
	fullFuncName := caller.Function
	parts := strings.Split(fullFuncName, "/")
	return parts[len(parts)-1]
}

func InitLogger() {
	Log = logrus.New()
	Log.SetFormatter(&CustomFormatter{})
	Log.SetOutput(os.Stdout)
	Log.SetReportCaller(true)

	// 根据配置设置日志级别
	switch strings.ToUpper(config.AppConfig.LogLevel) {
	case "DEBUG":
		Log.SetLevel(logrus.DebugLevel)
	case "INFO":
		Log.SetLevel(logrus.InfoLevel)
	case "WARN":
		Log.SetLevel(logrus.WarnLevel)
	case "ERROR":
		Log.SetLevel(logrus.ErrorLevel)
	case "FATAL":
		Log.SetLevel(logrus.FatalLevel)
	default:
		Log.SetLevel(logrus.InfoLevel) // 默认为 INFO
	}
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	Log.Debug(args...)
}

// Info logs an info message
func Info(args ...interface{}) {
	Log.Info(args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	Log.Warn(args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	Log.Error(args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	Log.Fatal(args...)
}
