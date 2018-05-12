package route

import (
	"net/http"
	"encoding/json"
)

type httpErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewHttpError(code int, message string) ([]byte) {
	resp := httpErrorResponse{code, message}
	b, _ := json.Marshal(&resp)
	return b
}

func ErrorResponse(w http.ResponseWriter, code int) {
	ErrorResponseWithMessage(w, code, http.StatusText(code))
}

func ErrorResponseWithMessage(w http.ResponseWriter, code int, message string) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	w.Write(NewHttpError(code, message))
}
