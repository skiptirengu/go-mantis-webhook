package route

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"log"
	"github.com/skiptirengu/go-mantis-webhook/config"
)

var Middleware = middleware{}

type middleware struct{}

func (m middleware) App(next httprouter.Handle) httprouter.Handle {
	return m.JSON(m.AuthorizeApplication(next))
}

func (middleware) JSON(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		next(w, r, p)
	}
}

func (middleware) AuthorizeWebhook(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if authorizeHeader(w, r, "X-Gitlab-Token", config.Get().Gitlab.Token) {
			next(w, r, p)
		}
	}
}

func (middleware) AuthorizeApplication(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if authorizeHeader(w, r, "Authorization", config.Get().Secret) {
			next(w, r, p)
		}
	}
}

func authorizeHeader(w http.ResponseWriter, r *http.Request, header, token string) (bool) {
	auth := r.Header.Get(header)
	if auth == "" || auth != token {
		log.Printf("Denied request on endpoint %s from host %s", r.URL, r.RemoteAddr)
		ErrorResponse(w, http.StatusUnauthorized)

		return false
	} else {
		return true
	}
}