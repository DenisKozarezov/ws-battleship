package logger

import (
	"log"
)

type DefaultLogger struct {
	isDebugMode bool
	logger      *log.Logger
}

func NewDefaultLogger(prefix string) *DefaultLogger {
	logger := log.New(log.Writer(), prefix+" ", log.LstdFlags)
	return &DefaultLogger{logger: logger}
}

func (l *DefaultLogger) Close() {}

func (l *DefaultLogger) SetDebugMode(mode bool) {
	l.isDebugMode = mode
}

func (l *DefaultLogger) Info(args ...any) {
	l.logger.Println(args...)
}

func (l *DefaultLogger) Infof(msg string, args ...any) {
	l.logger.Printf(msg, args...)
}

func (l *DefaultLogger) Fatal(args ...any) {
	l.logger.Fatalln(args...)
}

func (l *DefaultLogger) Fatalf(msg string, args ...any) {
	l.logger.Fatalf(msg, args...)
}

func (l *DefaultLogger) Error(args ...any) {
	l.Info(args...)
}

func (l *DefaultLogger) Errorf(msg string, args ...any) {
	l.Infof(msg, args...)
}

func (l *DefaultLogger) Debug(args ...any) {
	if l.isDebugMode {
		l.Info(args...)
	}
}

func (l *DefaultLogger) Debugf(msg string, args ...any) {
	if l.isDebugMode {
		l.Infof(msg, args...)
	}
}
