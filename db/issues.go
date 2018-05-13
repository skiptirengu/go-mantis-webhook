package db

import (
	"database/sql"
	"time"
	"github.com/lib/pq"
	"github.com/kisielk/sqlstruct"
)

var Issues = issues{Get()}

type issues struct {
	db *sql.DB
}

type IssuesTable struct {
	CommitHash string
	IssueID    int
	Email      string
	Date       time.Time
}

func (i issues) Close(issueID int, commitHash, email string) (error) {
	const insertSql = "insert into issues (commit_hash, issue_id, email) values ($1, $2, $3)"
	_, err := i.db.Exec(insertSql, commitHash, issueID, NullableStr(email))
	return err
}

func (i issues) Closed(commits []string) (map[string]*IssuesTable, error) {
	rows, err := i.db.Query("select * from issues where commit_hash = any($1)", pq.Array(commits))
	if err != nil {
		return nil, err
	}

	res := make(map[string]*IssuesTable, 0)
	for rows.Next() {
		issue := &IssuesTable{}
		sqlstruct.Scan(issue, rows)
		res[issue.CommitHash] = issue
	}

	return res, nil
}
