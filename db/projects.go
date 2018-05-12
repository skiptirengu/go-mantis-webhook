package db

import (
	"database/sql"
)

var Projects = projects{Get()}

type projects struct {
	db *sql.DB
}

type projectsTable struct {
	Mantis string `json:"mantis"`
	Gitlab string `json:"gitlab"`
	ID     int    `json:"id"`
}

func (p *projects) CheckExists(mantis, gitlab string) (bool) {
	var count int
	res, _ := p.db.Query("select count(*) from projects where mantis = $1 or gitlab = $2", mantis, gitlab)
	ScanCol(res, &count)
	return count > 0
}

func (p *projects) Create(mantis, gitlab string) (*projectsTable, error) {
	var insertedId int
	if res, err := p.db.Query("insert into projects (mantis, gitlab) values ($1, $2) returning id", mantis, gitlab); err != nil {
		return nil, err
	} else {
		ScanCol(res, &insertedId)
		return &projectsTable{ID: insertedId, Gitlab: gitlab, Mantis: mantis}, nil
	}
}
