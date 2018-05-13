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
	"github.com/skiptirengu/go-mantis-webhook/util"
	"strconv"
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

	// TODO use the api to get all commits using before and after refs
	if event.TotalCommitsCount > pushEventMaxCommits {
		log.Printf("This push have %d commits, processing only the first %d commits", event.TotalCommitsCount, pushEventMaxCommits)
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

		issues, err := hook.extractIssues(event.Commits)
		if err != nil {
			log.Println(err)
			return
		}
		if len(issues) == 0 {
			return
		}

		var (
			closedIssues = make([]int, 0)
			userCache    = make(map[string]*db.UsersTable, len(issues))
		)
		for email, issue := range issues {
			var (
				user *db.UsersTable
				ok   bool
				err  error
			)

			if user, ok = userCache[email]; !ok {
				user, err = db.Users.Get(email)
				switch err {
				case db.UserNotFound:
					if !synced {
						mantis.SyncProjectUsers(project.Mantis)
						user, err = db.Users.Get(email)
						synced = true
					}
				}
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

			if err = mantis.Rest.CloseIssue(issue.ID); err != nil {
				log.Println(err)
				continue
			}

			if err = db.Issues.Close(issue.ID, issue.CommitHash, user.Email); err != nil {
				log.Println(err)
				continue
			}

			closedIssues = append(closedIssues, issue.ID)
		}

		if l := len(closedIssues); l > 0 {
			mapped := util.MapStringSlice(closedIssues, func(v interface{}) string { return strconv.Itoa(v.(int)) })
			log.Printf("Webhook call closed %d issues (%s)", l, strings.Join(mapped, ", "))
		}
	}()
}

func (hook webhook) extractIssues(c []commits) (map[string]*commitWithID, error) {
	var (
		issues      = make(map[string]*commitWithID, len(c))
		mapped      = util.MapStringSlice(c, func(val interface{}) string { return val.(commits).ID })
		closed, err = db.Issues.Closed(mapped)
	)

	if err != nil {
		return nil, err
	}

	for _, commit := range c {
		// Skip already closed issues
		if _, ok := closed[commit.ID]; ok {
			continue
		}
		// Scan all issues closed on this commit
		for _, strIssueId := range commitRegex.FindAllString(commit.Message, -1) {
			// The regex matches the issue id prefixed with #
			strIssueId = strings.Replace(strIssueId, "#", "", -1)
			// Being unable to parse the int value means our regex is probably bugged :p
			intIssueId, err := strconv.Atoi(strIssueId)
			if err != nil {
				log.Printf("Cannot convert string(%s) to int wrong regex match?", strIssueId)
				continue
			}
			issues[commit.Author.Email] = &commitWithID{intIssueId, commit.ID}
		}
	}

	return issues, nil
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

type commitWithID struct {
	ID         int
	CommitHash string
}

type pushEvent struct {
	Before            string    `json:"before"`
	After             string    `json:"after"`
	Ref               string    `json:"ref"`
	Project           project   `json:"project"`
	Commits           []commits `json:"commits"`
	TotalCommitsCount int       `json:"total_commits_count"`
}

type commits struct {
	ID        string    `json:"id"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
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
	PathWithNamespace string `json:"path_with_namespace"`
}
