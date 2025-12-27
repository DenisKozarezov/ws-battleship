package logger

import (
	"fmt"
	"os"
	"path/filepath"
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

func NewLogger(isDebugMode bool, prefix string) (Logger, error) {
	if isDebugMode {
		return NewDefaultLogger(os.Stdout, prefix, Debug), nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}
	tempDir := filepath.Join(dir, tempFolderName)

	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		if err = os.Mkdir(tempDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create Temp folder: %w", err)
		}
	}

	logPath := filepath.Join(tempDir, logFileName)
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to create Log.txt file: %w", err)
	}

	return NewDefaultLogger(logFile, prefix, Info), nil
}
