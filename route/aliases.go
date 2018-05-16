package route

import (
	"encoding/json"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/skiptirengu/go-mantis-webhook/db"
)

type aliases struct {
	db db.Database
}

type addAliasRequest struct {
	Email string `json:"email"`
	Alias string `json:"alias"`
}

func Aliases(db db.Database) (*aliases) {
	return &aliases{db}
}

func (a aliases) Add(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var (
		request  = addAliasRequest{}
		database = a.db.Aliases()
	)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		ErrorResponse.Send(w, http.StatusInternalServerError)
		return
	}

	if database.CheckExist(request.Email) {
		ErrorResponse.SendWithMessage(w, http.StatusBadRequest, "An alias for this email already exists")
		return
	}

	res, err := database.Create(request.Email, request.Alias)
	DataResponse.Send(w, res, err)
}
