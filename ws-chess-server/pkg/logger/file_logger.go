package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const (
	logFileName    = "Log.txt"
	tempFolderName = "Temp"
)

type fileWriter interface {
	io.WriteCloser
	io.StringWriter
}

type FileLogger struct {
	logger      fileWriter
	prefix      string
	isDebugMode bool
}

func NewFileLogger(prefix string) *FileLogger {
	if _, err := os.Stat(tempFolderName); os.IsNotExist(err) {
		_ = os.Mkdir(tempFolderName, 0755)
	}

	tempFile, err := os.OpenFile(tempFolderName+"/"+logFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	_, _ = tempFile.WriteString("===================== BEGIN =====================\n")

	return &FileLogger{
		logger: tempFile,
		prefix: prefix,
	}
}

func (l *FileLogger) Close() error {
	_, _ = l.WriteString("===================== END =====================\n")
	return l.logger.Close()
}

func (l *FileLogger) SetDebugMode(mode bool) {
	l.isDebugMode = mode
}

func (l *FileLogger) Info(args ...any) {
	_, _ = l.WriteString(l.str(args...))
}

func (l *FileLogger) Infof(msg string, args ...any) {
	_, _ = l.WriteString(l.strf(msg+"\n", args...))
}

func (l *FileLogger) Fatal(args ...any) {
	l.Error(args...)
	panic(fmt.Sprintln(args...))
}

func (l *FileLogger) Fatalf(msg string, args ...any) {
	l.Errorf(msg, args...)
	panic(fmt.Sprintf(msg+"\n", args...))
}

func (l *FileLogger) Error(args ...any) {
	l.Info(args...)
}

func (l *FileLogger) Errorf(msg string, args ...any) {
	l.Infof(msg, args...)
}

func (l *FileLogger) Debug(args ...any) {
	if l.isDebugMode {
		l.Info(args...)
	}
}

func (l *FileLogger) Debugf(msg string, args ...any) {
	if l.isDebugMode {
		l.Infof(msg, args...)
	}
}

func (l *FileLogger) WriteString(str string) (n int, err error) {
	return l.logger.WriteString(str)
}

func (l *FileLogger) str(args ...any) string {
	datetime := time.Now().Format(time.DateTime)

	var builder strings.Builder
	builder.WriteString(l.prefix)
	builder.WriteRune(' ')
	builder.WriteString(datetime)
	builder.WriteRune(' ')
	builder.WriteString(fmt.Sprintln(args...))
	return builder.String()
}

func (l *FileLogger) strf(msg string, args ...any) string {
	datetime := time.Now().Format(time.DateTime)

	var builder strings.Builder
	builder.WriteString(l.prefix)
	builder.WriteRune(' ')
	builder.WriteString(datetime)
	builder.WriteRune(' ')
	builder.WriteString(fmt.Sprintf(msg, args...))
	return builder.String()
}
