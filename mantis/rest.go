package mantis

import (
	"fmt"
	"github.com/skiptirengu/go-mantis-webhook/config"
	"github.com/parnurzeal/gorequest"
	"strconv"
	"github.com/skiptirengu/go-mantis-webhook/util"
	"net/http"
	"github.com/pkg/errors"
	"strings"
	"encoding/json"
)

const closedIssueStatusID = 80

type rest struct {
	conf *config.Configs
}

var Rest = rest{config.Get()}

func (r rest) restAction(method string, params ...string) (string) {
	action := fmt.Sprintf("%s/%s", r.restEndpoint(), method)
	for _, param := range params {
		action += fmt.Sprintf("/%s", param)
	}
	return action
}

func (r rest) restEndpoint() (string) {
	return fmt.Sprintf("%s/api/rest", getHost())
}

func (r rest) CloseIssue(id int, userID int) (error) {
	var request = &closeIssueRequest{
		Status:  closeIssueRequestStatus{closedIssueStatusID},
		Handler: closeIssueRequestHandler{userID},
	}
	return r.makeCloseIssueRequest(id, request)
}

func (r rest) makeCloseIssueRequest(id int, request *closeIssueRequest) (error) {
	var action = r.restAction("issues", strconv.Itoa(id))
	_, err := r.makeRequest("PATCH", action, request)
	return err
}

func (r rest) makeRequest(method, action string, body interface{}, response ...interface{}) (*http.Response, error) {
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

	if err = util.ShiftError(errs); err != nil {
		return nil, err
	} else if resp.StatusCode >= 400 {
		return nil, errors.New(http.StatusText(resp.StatusCode))
	}

	if len(response) > 0 {
		err = json.NewDecoder(resp.Body).Decode(response[0])
	}

	return resp, err
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
