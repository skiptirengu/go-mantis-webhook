package route

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"log"
	"github.com/skiptirengu/go-mantis-webhook/config"
)

func AuthorizeWebhook(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if authorizeHeader(w, r, "X-Gitlab-Token", config.Get().Gitlab.Token) {
			next(w, r, p)
		}
	}
}

func AuthorizeApplication(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if authorizeHeader(w, r, "Authorization", config.Get().Secret) {
			next(w, r, p)
		}
	}
}

func authorizeHeader(w http.ResponseWriter, r *http.Request, token, header string) (bool) {
	auth := r.Header.Get(header)
	if auth == "" || auth != token {
		log.Printf("Denied request on endpoint %s from host %s", r.URL, r.RemoteAddr)
		w.WriteHeader(http.StatusUnauthorized)

		return false
	} else {
		return true
	}
}
