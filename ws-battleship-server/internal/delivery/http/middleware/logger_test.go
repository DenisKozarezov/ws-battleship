package middleware

import (
	"net/http"
	"net/url"
	"testing"
	"time"
)

func BenchmarkRequestLog(b *testing.B) {
	baseUrl, _ := url.Parse("/ws")
	r := http.Request{
		Method:     http.MethodGet,
		URL:        baseUrl,
		RemoteAddr: "127.0.0.1:8080",
	}
	elapsed := time.Second * 2

	b.ResetTimer()
	for b.Loop() {
		makePreAllocatedLog(&r, elapsed)
	}
}

func TestPreAllocatedLog(t *testing.T) {
	for _, tt := range []struct {
		name     string
		request  http.Request
		url      string
		elapsed  time.Duration
		expected string
	}{
		{
			name: "log with seconds",
			request: http.Request{
				RemoteAddr: "127.0.0.1:8080",
				Method:     http.MethodGet,
			},
			url:      "/api/v1/route",
			elapsed:  time.Second * 2,
			expected: `[GET]  |  "/api/v1/route"  |  2.000s  |  127.0.0.1:8080`,
		},
		{
			name: "log with milliseconds",
			request: http.Request{
				RemoteAddr: "127.0.0.1:8080",
				Method:     http.MethodPatch,
			},
			url:      "/api/v1/route?arg1=123&arg2=abc",
			elapsed:  time.Millisecond * 123,
			expected: `[PATCH]  |  "/api/v1/route?arg1=123&arg2=abc"  |  0.123s  |  127.0.0.1:8080`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Arrange
			tt.request.URL, _ = url.Parse(tt.url)

			// 2. Act
			got := makePreAllocatedLog(&tt.request, tt.elapsed)

			// 3. Assert
			if tt.expected != got {
				t.Logf("expected:\t'%s'", tt.expected)
				t.Logf("got:\t\t'%s'", got)
				t.Fail()
			}
		})
	}
}
