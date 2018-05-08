package main

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"log"
)

func AuthMiddleware(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		reqToken := r.Header.Get("X-Gitlab-Token")
		if reqToken == "" || reqToken != GetConfig().Token {
			log.Printf("Denied webhook request on endpoint %s from host %s", r.URL, r.RemoteAddr)
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			next(w, r, p)
		}
	}
}
