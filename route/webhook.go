package route

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"regexp"
	"time"
	"encoding/json"
	"log"
)

var (
	// TODO fix this regex and it's capturing groups
	// The regex have three parts
	// First part matches the verbs: Fix, fixed, closed, closes, closing, resolve, resolved, etc
	// Second part matches the task id: #1, #12, #19 #12, etc
	// Last part matches the time (if any): 0.5, 1, 1.2, etc
	commitRegex, _ = regexp.Compile("(?:[Cc]los(?:e[sd]?|ing)|[Ff]ix(?:e[sd]|ing)?|[Rr]esolv(?:e[sd]?|ing)|[Ii]mplement(?:s|ed|ing)?)?(?:[:\\s])*((?:#)\\d+(?:[\\s,])*)+$")
)

var Webhook = webhook{}

type webhook struct{}

func (webhook) Push(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	event := &pushEvent{}

	if err := json.NewDecoder(r.Body).Decode(event); err != nil {
		log.Print("Unable to parse webhook body", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO actual implementation
	return
}

type pushEvent struct {
	ObjectKind        string     `json:"object_kind"`
	Before            string     `json:"before"`
	After             string     `json:"after"`
	Ref               string     `json:"ref"`
	CheckoutSha       string     `json:"checkout_sha"`
	UserID            int        `json:"user_id"`
	UserName          string     `json:"user_name"`
	UserUsername      string     `json:"user_username"`
	UserEmail         string     `json:"user_email"`
	UserAvatar        string     `json:"user_avatar"`
	ProjectID         int        `json:"project_id"`
	Project           project    `json:"project"`
	Repository        repository `json:"repository"`
	Commits           []commits  `json:"commits"`
	TotalCommitsCount int        `json:"total_commits_count"`
}

type repository struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	Description     string `json:"description"`
	Homepage        string `json:"homepage"`
	GitHTTPURL      string `json:"git_http_url"`
	GitSSHURL       string `json:"git_ssh_url"`
	VisibilityLevel int    `json:"visibility_level"`
}

type commits struct {
	ID        string    `json:"id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	URL       string    `json:"url"`
	Author    author    `json:"author"`
}

type author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type project struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	WebURL            string `json:"web_url"`
	Namespace         string `json:"namespace"`
	VisibilityLevel   int    `json:"visibility_level"`
	PathWithNamespace string `json:"path_with_namespace"`
	DefaultBranch     string `json:"default_branch"`
	Homepage          string `json:"homepage"`
	URL               string `json:"url"`
}
