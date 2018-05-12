package route

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"encoding/json"
	"github.com/skiptirengu/go-mantis-webhook/db"
)

var Projects = projects{}

type projects struct{}

type addProjectRequest struct {
	GitlabProject string `json:"gitlab_project"`
	MantisProject string `json:"mantis_project"`
}

func (projects) Add(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var request = addProjectRequest{}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		ErrorResponse.Send(w, http.StatusInternalServerError)
		return
	}

	if db.Projects.CheckExists(request.MantisProject, request.GitlabProject) {
		ErrorResponse.SendWithMessage(w, http.StatusBadRequest, "Project already exists")
		return
	}

	res, err := db.Projects.Create(request.MantisProject, request.GitlabProject)
	DataResponse.Send(w, res, err)
}
