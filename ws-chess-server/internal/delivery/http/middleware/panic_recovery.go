package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"ws-chess-server/internal/delivery/http/response"
)

func PanicRecovery(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				response.Error(w, errors.New(fmt.Sprint(err)), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	}
}
