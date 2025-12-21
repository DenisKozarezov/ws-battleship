package middleware

import (
	"net/http"
	"strconv"
	"time"
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
}

const (
	separator = "  |  "
)

func LoggerMiddleware(logger Logger, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		next.ServeHTTP(w, r)
		elapsed := time.Since(now)

		logger.Info(makePreAllocatedLog(r, elapsed))
	}
}

func makePreAllocatedLog(r *http.Request, elapsed time.Duration) string {
	const separatorsLen = 3 * len(separator)

	urlLen := len(r.URL.Path)
	if r.URL.RawQuery != "" {
		urlLen += 1 + len(r.URL.RawQuery)
	}

	timeBuf := make([]byte, 0, 32)
	timeBuf = strconv.AppendFloat(timeBuf, elapsed.Seconds(), 'f', 3, 64)

	totalSize := 2 + len(r.Method) + // "[" + GET + "]"
		separatorsLen + // "   |   "
		2 + urlLen + // `"` + url + `"`
		len(timeBuf) + 1 + // 0.000s
		len(r.RemoteAddr) // 127.0.0.1:8080

	buf := make([]byte, 0, totalSize)

	// Print method. Example: "[GET]"
	buf = append(buf, '[')
	buf = append(buf, r.Method...)
	buf = append(buf, ']')

	// Print full URL path. Example: "/api/v1/my-route?arg1=123&arg2=abc"
	buf = append(buf, separator...)
	buf = append(buf, '"')
	buf = append(buf, r.URL.Path...)
	if r.URL.RawQuery != "" {
		buf = append(buf, '?')
		buf = append(buf, r.URL.RawQuery...)
	}
	buf = append(buf, '"')

	// Print elapsed duration of the request. Example: "0.000s"
	buf = append(buf, separator...)
	buf = append(buf, timeBuf...)
	buf = append(buf, 's')

	// Print remote IP-Address. Example: "127.0.0.1:8080"
	buf = append(buf, separator...)
	buf = append(buf, r.RemoteAddr...)

	return string(buf)
}
