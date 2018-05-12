package route

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"encoding/json"
	"github.com/skiptirengu/go-mantis-webhook/db"
	"log"
)

type AddProjectRequest struct {
	GitlabProject string `json:"gitlab_project"`
	MantisProject string `json:"mantis_project"`
}

var Projects = projects{}

type projects struct{}

func (projects) Add(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var request = AddProjectRequest{}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		ErrorResponse(w, http.StatusInternalServerError)
		return
	}

	if db.Projects.CheckExists(request.MantisProject, request.GitlabProject) {
		ErrorResponseWithMessage(w, http.StatusBadRequest, "Project already exists")
		return
	}

	if res, err := db.Projects.Create(request.MantisProject, request.GitlabProject); err == nil {
		w.WriteHeader(http.StatusCreated)
		data, _ := json.Marshal(res)
		w.Write(data)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err)
	}
}
