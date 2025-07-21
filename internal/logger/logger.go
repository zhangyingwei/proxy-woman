package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// LogLevel 日志级别
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Logger 日志记录器
type Logger struct {
	level   LogLevel
	logFile *os.File
}

var defaultLogger *Logger

// InitLogger 初始化日志记录器
func InitLogger(configDir string, level string) error {
	logLevel := parseLogLevel(level)
	
	// 创建日志目录
	logDir := filepath.Join(configDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// 创建日志文件
	logFileName := fmt.Sprintf("proxywoman_%s.log", time.Now().Format("2006-01-02"))
	logPath := filepath.Join(logDir, logFileName)
	
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	defaultLogger = &Logger{
		level:   logLevel,
		logFile: logFile,
	}

	// 设置标准日志输出到文件
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	return nil
}

// parseLogLevel 解析日志级别
func parseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}

// Debug 调试日志
func Debug(format string, args ...interface{}) {
	if defaultLogger != nil && defaultLogger.level <= DEBUG {
		log.Printf("[DEBUG] "+format, args...)
	}
}

// Info 信息日志
func Info(format string, args ...interface{}) {
	if defaultLogger != nil && defaultLogger.level <= INFO {
		log.Printf("[INFO] "+format, args...)
	}
}

// Warn 警告日志
func Warn(format string, args ...interface{}) {
	if defaultLogger != nil && defaultLogger.level <= WARN {
		log.Printf("[WARN] "+format, args...)
	}
}

// Error 错误日志
func Error(format string, args ...interface{}) {
	if defaultLogger != nil && defaultLogger.level <= ERROR {
		log.Printf("[ERROR] "+format, args...)
	}
}

// Close 关闭日志记录器
func Close() {
	if defaultLogger != nil && defaultLogger.logFile != nil {
		defaultLogger.logFile.Close()
	}
}
