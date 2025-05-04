package utils

import (
	"log"
	"os"
)

// Logger is a simple logger interface
type Logger interface {
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
	Debug(format string, args ...interface{})
}

// DefaultLogger is an implementation of Logger using standard log package
type DefaultLogger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
}

// NewDefaultLogger creates a new instance of DefaultLogger
func NewDefaultLogger() Logger {
	return &DefaultLogger{
		infoLogger:  log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLogger: log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

// Info logs an info message
func (l *DefaultLogger) Info(format string, args ...interface{}) {
	l.infoLogger.Printf(format, args...)
}

// Error logs an error message
func (l *DefaultLogger) Error(format string, args ...interface{}) {
	l.errorLogger.Printf(format, args...)
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(format string, args ...interface{}) {
	l.debugLogger.Printf(format, args...)
}
