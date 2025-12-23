package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Status int `json:"status"`
	Data   any `json:"data"`
}

type ErrorResponse struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
}

func Error(w http.ResponseWriter, err error, code int) {
	ResponseWithJSON(w, code, &ErrorResponse{
		Status: code,
		Error:  err.Error(),
	})
}

func ResponseWithStatus(w http.ResponseWriter, code int) {
	ResponseWithJSON(w, code, Response{Status: code})
}

func ResponseWithJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)

	_ = json.NewEncoder(w).Encode(data)
}
