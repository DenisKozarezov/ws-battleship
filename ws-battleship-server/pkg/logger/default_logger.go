package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

type Level int32

const (
	Debug Level = iota
	Info
	Error
	Fatal
)

func (l Level) String() string {
	switch l {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Error:
		return "ERROR"
	case Fatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

type DefaultLogger struct {
	level  Level
	logger *log.Logger
}

func NewDefaultLogger(out io.Writer, prefix string, lvl Level) *DefaultLogger {
	logger := log.New(out, prefix+" ", log.LstdFlags)
	return &DefaultLogger{
		logger: logger,
		level:  lvl,
	}
}

func (l *DefaultLogger) SetLevel(lvl Level) {
	l.level = lvl
}

func (l *DefaultLogger) Debug(args ...any) {
	l.logf(Debug, "%s", fmt.Sprint(args...))
}

func (l *DefaultLogger) Debugf(format string, args ...any) {
	l.logf(Debug, format, args...)
}

func (l *DefaultLogger) Info(args ...any) {
	l.logf(Info, "%s", fmt.Sprint(args...))
}

func (l *DefaultLogger) Infof(format string, args ...any) {
	l.logf(Info, format, args...)
}

func (l *DefaultLogger) Error(args ...any) {
	l.logf(Error, "%s", fmt.Sprint(args...))
}

func (l *DefaultLogger) Errorf(format string, args ...any) {
	l.logf(Error, format, args...)
}

func (l *DefaultLogger) Fatal(args ...any) {
	l.logf(Fatal, "%s", fmt.Sprint(args...))
}

func (l *DefaultLogger) Fatalf(format string, args ...any) {
	l.logf(Fatal, format, args...)
}

func (l *DefaultLogger) logf(level Level, msg string, args ...any) {
	if level < Level(l.level) {
		return
	}

	l.logger.Printf("[%s] "+msg, append([]any{level.String()}, args...)...)

	if level == Fatal {
		os.Exit(1)
	}
}
