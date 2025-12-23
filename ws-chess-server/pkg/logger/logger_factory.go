package logger

import (
	"io"
	"ws-chess-server/internal/config"
)

type Logger interface {
	io.Closer

	Info(args ...any)
	Infof(msg string, args ...any)
	Fatal(args ...any)
	Fatalf(msg string, args ...any)
	Error(args ...any)
	Errorf(msg string, args ...any)
	Debug(args ...any)
	Debugf(msg string, args ...any)
	SetDebugMode(isDebugMode bool)
}

func NewLogger(cfg *config.AppConfig, prefix string) Logger {
	var logger Logger

	if cfg.IsDebugMode {
		logger = NewDefaultLogger(prefix)
	} else {
		logger = NewFileLogger(prefix)
	}

	logger.SetDebugMode(cfg.IsDebugMode)

	return logger
}
