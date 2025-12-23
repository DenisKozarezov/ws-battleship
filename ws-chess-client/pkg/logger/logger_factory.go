package logger

import (
	"fmt"
	"os"
	"ws-chess-client/internal/config"
)

type Logger interface {
	Info(args ...any)
	Infof(msg string, args ...any)
	Fatal(args ...any)
	Fatalf(msg string, args ...any)
	Error(args ...any)
	Errorf(msg string, args ...any)
	Debug(args ...any)
	Debugf(msg string, args ...any)
	SetLevel(level Level)
}

const (
	tempFolderName = "Temp"
	logFileName    = "Log.txt"
)

func NewLogger(cfg *config.AppConfig, prefix string) (Logger, error) {
	if cfg.IsDebugMode {
		return NewDefaultLogger(os.Stdout, prefix, Debug), nil
	}

	if _, err := os.Stat(tempFolderName); os.IsNotExist(err) {
		if err = os.Mkdir(tempFolderName, 0755); err != nil {
			return nil, fmt.Errorf("failed to create Temp folder: %w", err)
		}
	}

	tempFile, err := os.OpenFile(tempFolderName+"/"+logFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to create Log.txt file: %w", err)
	}

	return NewDefaultLogger(tempFile, prefix, Info), nil
}
