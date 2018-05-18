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
	"github.com/skiptirengu/go-mantis-webhook/config"
	"fmt"
	"errors"
)

const pushEventMaxCommits = 20

var (
	// TODO fix this regex and it's capturing groups
	// (?:[Cc]los(?:e[sd]?|ing)|[Ff]ix(?:e[sd]|ing)?|[Rr]esolv(?:e[sd]?|ing)|[Ii]mplement(?:s|ed|ing)?)+(?:[:\s])*((?:#)*(\d+)(?:[,\s])*)+
	commitRegex, _ = regexp.Compile("(?m)#[1-9]\\d*")
)

type webhook struct {
	conf *config.Configuration
	db   db.Database
}

func Webhook(c *config.Configuration, db db.Database) (*webhook) {
	return &webhook{c, db}
}

func (h webhook) Push(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
		project, err := h.getProject(event.Project.PathWithNamespace)
		if err != nil {
			log.Println(err)
			return
		}

		issues, err := h.extractIssues(event.Commits)
		if err != nil {
			log.Println(err)
			return
		}
		if len(issues) == 0 {
			log.Println("No closeable commits found. Skipping...")
			return
		}

		var (
			restService  = mantis.NewRestService(h.conf)
			synced       = false
			closedIssues = make([]int, 0)
			userCache    = make(map[string]*db.UsersTable, len(issues))
		)
		for email, issue := range issues {
			var (
				user      *db.UsersTable
				ok        bool
				err       error
				userEmail string
			)

			if user, ok = userCache[email]; !ok {
				user, err = h.db.Users().Get(email)
				switch err {
				case db.UserNotFound:
					if !synced {
						mantis.SyncProjectUsers(h.conf, h.db, project.Mantis)
						user, err = h.db.Users().Get(email)
						synced = true
					} else {
						err = nil
					}
				}
			}

			if err != nil {
				log.Println(err)
				continue
			} else if user == nil {
				log.Printf("Unable to find user with email %s", email)
			} else {
				userEmail = user.Email
				userCache[email] = user
			}

			if user == nil {
				err = restService.CloseIssue(issue.ID, 0)
			} else {
				err = restService.CloseIssue(issue.ID, user.ID)
			}

			if err != nil {
				log.Println(err)
				continue
			}

			if err = h.db.Issues().Close(issue.ID, issue.CommitHash, userEmail); err != nil {
				log.Println(err)
			}

			var message = fmt.Sprintf("Tarefa fechada no commit %s", issue.URL)
			if user != nil {
				message += fmt.Sprintf(" pelo usuário %s.", user.Name)
			}

			if err = restService.AddNote(issue.ID, message); err != nil {
				log.Println(err)
			}

			closedIssues = append(closedIssues, issue.ID)
		}

		if l := len(closedIssues); l > 0 {
			mapped := util.MapStringSlice(closedIssues, func(v interface{}) string { return strconv.Itoa(v.(int)) })
			log.Printf("Webhook call closed %d issues (%s)", l, strings.Join(mapped, ", "))
		}
	}()
}

func (h webhook) extractIssues(c []commits) (map[string]*commitWithID, error) {
	var (
		issues      = make(map[string]*commitWithID, len(c))
		mapped      = util.MapStringSlice(c, func(val interface{}) string { return val.(commits).ID })
		closed, err = h.db.Issues().Closed(mapped)
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
				log.Printf("Cannot convert string(%s) to int, wrong regex match?", strIssueId)
				continue
			}
			issues[commit.Author.Email] = &commitWithID{intIssueId, commit.ID, commit.URL}
		}
	}

	return issues, nil
}

func (h webhook) getProject(projectWithNamespace string) (*db.ProjectsTable, error) {
	if p, err := h.db.Projects().Get(projectWithNamespace); err != nil {
		switch err {
		case db.ProjectNotFound:
			return nil, errors.New(fmt.Sprintf("Unable to find a mantis project vinculated to the gilab project \"%s\"", projectWithNamespace))
		default:
			return nil, err
		}
	} else {
		return p, nil
	}
}

type commitWithID struct {
	ID         int
	CommitHash string
	URL        string
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
	URL       string    `json:"url"`
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
