package route

import (
	"encoding/json"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/skiptirengu/go-mantis-webhook/db"
)

var Aliases = aliases{}

type aliases struct{}

type addAliasRequest struct {
	Email string `json:"email"`
	Alias string `json:"alias"`
}

func (*aliases) Add(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var request = addAliasRequest{}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		ErrorResponse.Send(w, http.StatusInternalServerError)
		return
	}

	if db.Aliases.CheckExist(request.Alias) {
		ErrorResponse.SendWithMessage(w, http.StatusBadRequest, "Alias already exists")
		return
	}

	res, err := db.Aliases.Create(request.Email, request.Alias)
	DataResponse.Send(w, res, err)
}
