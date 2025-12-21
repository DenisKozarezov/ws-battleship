package middleware

type Logger interface {
	Info(args ...any)
	Infof(msg string, args ...any)
	Fatal(args ...any)
	Fatalf(msg string, args ...any)
	Error(args ...any)
	Errorf(msg string, args ...any)
	Debug(args ...any)
	Debugf(msg string, args ...any)
}
