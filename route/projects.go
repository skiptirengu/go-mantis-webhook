package route

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"encoding/json"
	"github.com/skiptirengu/go-mantis-webhook/db"
)

type projects struct {
	db db.Database
}

type addProjectRequest struct {
	GitlabProject string `json:"gitlab_project"`
	MantisProject string `json:"mantis_project"`
}

func Projects(db db.Database) (*projects) {
	return &projects{db}
}

func (p projects) Add(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var (
		request  = addProjectRequest{}
		database = p.db.Projects()
	)

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		ErrorResponse.Send(w, http.StatusInternalServerError)
		return
	}

	if database.CheckExists(request.GitlabProject) {
		ErrorResponse.SendWithMessage(w, http.StatusBadRequest, "Project already exists")
		return
	}

	res, err := database.Create(request.MantisProject, request.GitlabProject)
	DataResponse.Send(w, res, err)
}
