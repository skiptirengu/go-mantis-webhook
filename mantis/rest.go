package mantis

import (
	"fmt"
	"github.com/skiptirengu/go-mantis-webhook/config"
	"github.com/parnurzeal/gorequest"
	"strconv"
	"github.com/skiptirengu/go-mantis-webhook/util"
	"net/http"
	"strings"
	"encoding/json"
	"errors"
)

const closedIssueStatusID = 80

type Rest struct {
	conf *config.Configuration
}

type addIssueNoteRequest struct {
	Text string `json:"text"`
}

type closeIssueRequest struct {
	Status  closeIssueRequestStatus  `json:"status"`
	Handler closeIssueRequestHandler `json:"handler,omitempty"`
}

type closeIssueRequestHandler struct {
	ID int `json:"id,omitempty"`
}

type closeIssueRequestStatus struct {
	ID int `json:"id"`
}

func NewRestService(c *config.Configuration) (*Rest) {
	return &Rest{c}
}

func (r Rest) AddNote(id int, note string) (error) {
	var request = &addIssueNoteRequest{
		Text: note,
	}
	return r.makeAddIssueNoteRequest(id, request)
}

func (r Rest) CloseIssue(id int, userID int) (error) {
	var request = &closeIssueRequest{
		Status:  closeIssueRequestStatus{closedIssueStatusID},
		Handler: closeIssueRequestHandler{userID},
	}
	return r.makeCloseIssueRequest(id, request)
}

func (r Rest) restAction(method string, params ...string) (string) {
	action := fmt.Sprintf("%s/%s", r.restEndpoint(), method)
	for _, param := range params {
		action += fmt.Sprintf("/%s", param)
	}
	return action
}

func (r Rest) restEndpoint() (string) {
	return fmt.Sprintf("%s/api/rest", getHost(r.conf))
}

func (r Rest) makeAddIssueNoteRequest(id int, request *addIssueNoteRequest) (error) {
	var action = fmt.Sprintf("%s/notes", r.restAction("issues", strconv.Itoa(id)))
	_, err := r.makeRequest("POST", action, request)
	return err
}

func (r Rest) makeCloseIssueRequest(id int, request *closeIssueRequest) (error) {
	var action = r.restAction("issues", strconv.Itoa(id))
	_, err := r.makeRequest("PATCH", action, request)
	return err
}

func (r Rest) makeRequest(method, action string, body interface{}, response ...interface{}) (*http.Response, error) {
	var (
		req = gorequest.New()
		err error
	)

	switch strings.ToLower(method) {
	case "post":
		req.Post(action)
	case "put":
		req.Put(action)
	case "patch":
		req.Patch(action)
	default:
		req.Get(action)
	}

	req.AppendHeader("Authorization", r.conf.Mantis.Token)

	if body != nil {
		req.SendStruct(body)
	}

	resp, _, errs := req.End()

	if err = util.PopError(errs); err != nil {
		return nil, err
	} else if resp.StatusCode >= 400 {
		return nil, errors.New(http.StatusText(resp.StatusCode))
	}

	if len(response) > 0 {
		err = json.NewDecoder(resp.Body).Decode(response[0])
	}

	return resp, err
}
