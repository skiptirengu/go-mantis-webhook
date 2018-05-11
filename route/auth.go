package route

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"log"
	"github.com/skiptirengu/go-mantis-webhook/config"
)

func AuthMiddleware(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		reqToken := r.Header.Get("X-Gitlab-Token")
		if reqToken == "" || reqToken != config.Get().Token {
			log.Printf("Denied webhook request on endpoint %s from host %s", r.URL, r.RemoteAddr)
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			next(w, r, p)
		}
	}
}
