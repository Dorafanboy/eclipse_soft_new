package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Level int

const (
	DEBUG Level = iota
	INFO
	SUCCESS
	WARNING
	ERROR
)

var levelColors = map[Level]string{
	DEBUG:   "\033[37m", // White
	INFO:    "\033[34m", // Blue
	SUCCESS: "\033[32m", // Green
	WARNING: "\033[33m", // Yellow
	ERROR:   "\033[31m", // Red
}

type Logger struct {
	logger *log.Logger
}

var defaultLogger *Logger

func init() {
	defaultLogger = New()
}

func New() *Logger {
	return &Logger{
		logger: log.New(os.Stdout, "", 0),
	}
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	color := levelColors[level]
	reset := "\033[0m"
	prefix := fmt.Sprintf("%s[%s]%s ", color, level.String(), reset)
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	message := fmt.Sprintf(format, args...)
	l.logger.Printf("%s %s%s", timestamp, prefix, message)
}

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case SUCCESS:
		return "SUCCESS"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func Debug(format string, args ...interface{}) {
	defaultLogger.log(DEBUG, format, args...)
}

func Info(format string, args ...interface{}) {
	defaultLogger.log(INFO, format, args...)
}

func Success(format string, args ...interface{}) {
	defaultLogger.log(SUCCESS, format, args...)
}

func Warning(format string, args ...interface{}) {
	defaultLogger.log(WARNING, format, args...)
}

func Error(format string, args ...interface{}) {
	defaultLogger.log(ERROR, format, args...)
}
