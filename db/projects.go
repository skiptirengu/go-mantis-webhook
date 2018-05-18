package db

import (
	"database/sql"
	"github.com/kisielk/sqlstruct"
	"errors"
)

var ProjectNotFound = errors.New("project not found")

type projects struct {
	db *sql.DB
}

type ProjectsTable struct {
	Mantis string `json:"mantis"`
	Gitlab string `json:"gitlab"`
	ID     int    `json:"id"`
}

func (p projects) Get(gitlab string) (*ProjectsTable, error) {
	rows, err := p.db.Query("select * from projects where gitlab = $1", gitlab)

	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, ProjectNotFound
	}

	res := &ProjectsTable{}
	sqlstruct.Scan(res, rows)

	return res, nil
}

func (p projects) CheckExists(mantis, gitlab string) (bool) {
	var count int
	res, _ := p.db.Query("select count(*) from projects where mantis = $1 or gitlab = $2", mantis, gitlab)
	ScanCol(res, &count)
	return count > 0
}

func (p projects) Create(mantis, gitlab string) (*ProjectsTable, error) {
	var insertedId int
	if res, err := p.db.Query("insert into projects (mantis, gitlab) values ($1, $2) returning id", mantis, gitlab); err != nil {
		return nil, err
	} else {
		ScanCol(res, &insertedId)
		return &ProjectsTable{ID: insertedId, Gitlab: gitlab, Mantis: mantis}, nil
	}
}
