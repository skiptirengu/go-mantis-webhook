package route

import (
	"net/http"
	"encoding/json"
	"log"
)

type httpErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var ErrorResponse = errorResponse{}
var DataResponse = dataResponse{}

type dataResponse struct{}
type errorResponse struct{}

func NewHttpError(code int, message string) ([]byte) {
	resp := httpErrorResponse{code, message}
	b, _ := json.Marshal(&resp)
	return b
}

func (e errorResponse) Send(w http.ResponseWriter, code int) {
	e.SendWithMessage(w, code, http.StatusText(code))
}

func (errorResponse) SendWithMessage(w http.ResponseWriter, code int, message string) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	w.Write(NewHttpError(code, message))
}

func (dataResponse) Send(w http.ResponseWriter, res interface{}, err error) {
	if err == nil {
		data, _ := json.Marshal(res)
		w.Write(data)
	} else {
		ErrorResponse.Send(w, http.StatusInternalServerError)
		log.Print(err)
	}
}
