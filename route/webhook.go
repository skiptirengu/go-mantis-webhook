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
	commitRegex, _ = regexp.Compile("#(\\d+)[^\\d](\\d+(\\.\\d+)?)")
)

type webhook struct {
	conf        *config.Configuration
	db          db.Database
	restService *mantis.Rest
}

func Webhook(c *config.Configuration, db db.Database) (*webhook) {
	return &webhook{c, db, mantis.NewRestService(c)}
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
		var issuesLen int

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
		if issuesLen = len(issues); issuesLen == 0 {
			log.Println("No closeable commits found. Skipping...")
			return
		}
		if err := mantis.SyncProjectUsers(h.conf, h.db, project.Mantis); err != nil {
			log.Println(err)
		}

		var (
			okChan       = make(chan int)
			errChan      = make(chan error)
			userCache    = make(map[string]*db.UsersTable, issuesLen)
			closedIssues = make([]int, 0)
		)

		for _, issue := range issues {
			var (
				user  *db.UsersTable
				ok    bool
				err   error
				email = issue.Email
			)

			if user, ok = userCache[email]; !ok {
				if user, err = h.db.Users().Get(email); err == nil {
					userCache[email] = user
				} else {
					log.Println(err)
				}
			}

			go h.closeIssue(okChan, errChan, issue, user)
		}

		for counter := issuesLen; counter > 0; counter-- {
			select {
			case id := <-okChan:
				closedIssues = append(closedIssues, id)
			case err := <-errChan:
				log.Println(err)
			}
		}

		if l := len(closedIssues); l > 0 {
			mapped := util.MapStringSlice(closedIssues, func(v interface{}) string { return strconv.Itoa(v.(int)) })
			log.Printf("Webhook call closed %d issues (%s)", l, strings.Join(mapped, ", "))
		}
	}()
}

func (h webhook) closeIssue(okChan chan int, errChan chan error, issue *parsedCommit, user *db.UsersTable) {
	var (
		err       error
		message   string
		userEmail string
	)

	if user == nil {
		err = h.restService.CloseIssue(issue.ID, 0, issue.hours)
	} else {
		userEmail = user.Email
		err = h.restService.CloseIssue(issue.ID, user.ID, issue.hours)
	}

	if err != nil {
		errChan <- err
		return
	}

	message = fmt.Sprintf("Tarefa fechada no commit %s", issue.URL)
	if user != nil {
		message += fmt.Sprintf(" pelo usuário %s.", user.Name)
	}

	err = h.db.Issues().Close(issue.ID, issue.CommitHash, userEmail)
	err = h.restService.AddNote(issue.ID, message)

	if err != nil {
		errChan <- err
	} else {
		okChan <- issue.ID
	}
}

func (h webhook) extractIssues(c []commits) ([]*parsedCommit, error) {
	var (
		issues      = make([]*parsedCommit, 0)
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
		for _, match := range commitRegex.FindAllStringSubmatch(commit.Message, -1) {
			if len(match) != 4 {
				continue
			}
			var (
				err       error
				issueId   int
				timeSpent float64
			)
			issueId, err = strconv.Atoi(match[1])
			timeSpent, err = strconv.ParseFloat(match[2], 32)
			// Being unable to parse the numbers value means our regex is probably bugged
			if err != nil {
				log.Println(err)
				continue
			}
			issues = append(issues, &parsedCommit{
				ID:         issueId,
				CommitHash: commit.ID,
				URL:        commit.URL,
				hours:      timeSpent,
				Email:      commit.Author.Email,
			})
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

type parsedCommit struct {
	ID         int
	CommitHash string
	URL        string
	hours      float64
	Email      string
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
