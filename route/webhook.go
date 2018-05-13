package route

import (
	"net/http"
	"github.com/julienschmidt/httprouter"
	"regexp"
	"time"
	"encoding/json"
	"log"
	"github.com/skiptirengu/go-mantis-webhook/db"
	"strings"
	"github.com/skiptirengu/go-mantis-webhook/mantis"
)

const pushEventMaxCommits = 20

var (
	// TODO fix this regex and it's capturing groups
	// (?:[Cc]los(?:e[sd]?|ing)|[Ff]ix(?:e[sd]|ing)?|[Rr]esolv(?:e[sd]?|ing)|[Ii]mplement(?:s|ed|ing)?)+(?:[:\s])*((?:#)*(\d+)(?:[,\s])*)+
	commitRegex, _ = regexp.Compile("(?m)#[1-9]\\d*")
)

var Webhook = webhook{}

type webhook struct{}

func (hook webhook) Push(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var event = pushEvent{}

	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		log.Print("Unable to parse webhook body", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	// Do all processing on background
	go func() {
		var synced = false

		project, err := hook.getProject(event.Project.PathWithNamespace)
		if err != nil {
			log.Println(err)
			return
		}

		issues := hook.extractIssues(event.Commits)
		if len(issues) == 0 {
			return
		}

		userCache := make(map[string]*db.UsersTable, len(issues))
		for email := range issues {
			var (
				user *db.UsersTable
				ok   bool
				err  error
			)

			if user, ok = userCache[email]; !ok {
				user, err = db.Users.Get(email)
			} else if !synced {
				mantis.SyncProjectUsers(project.Mantis)
				synced = true
				user, err = db.Users.Get(email)
			}

			if err != nil {
				log.Println(err)
				continue
			} else if user == nil {
				log.Printf("Unable to find user with email %s", email)
				continue
			} else {
				userCache[email] = user
			}

			//TODO close issue
		}
	}()
}

func (webhook) extractIssues(commits []commits) (map[string]string) {
	var issues = make(map[string]string, len(commits))
	for _, commit := range commits {
		for _, issueId := range commitRegex.FindAllString(commit.Message, -1) {
			issueId = strings.Replace(issueId, "#", "", -1)
			issues[commit.Author.Email] = issueId
		}
	}
	return issues
}

func (hook webhook) getProject(projectWithNamespace string) (*db.ProjectsTable, error) {
	if p, err := db.Projects.Get(projectWithNamespace); err != nil {
		switch err {
		case db.ProjectNotFound:
			log.Printf("Unable to find a mantis project vinculated to the gilab project \"%s\"", projectWithNamespace)
			return nil, err
		default:
			log.Println(err)
			return nil, err
		}
	} else {
		return p, nil
	}
}

func (hook webhook) tryImportUser() () {

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
