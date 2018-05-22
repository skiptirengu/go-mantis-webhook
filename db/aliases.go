package db

import (
	"database/sql"
)

type aliases struct {
	db *sql.DB
}

type aliasesTable struct {
	Email string `json:"email"`
	Alias string `json:"alias"`
}

func (a aliases) CheckExist(email string) (bool) {
	var count int
	res, err := a.db.Query("select count(*) from aliases where email = $1", email)
	if err == nil {
		defer res.Close()
	}
	ScanCol(res, &count)
	return count > 0
}

func (a aliases) Create(email string, alias string) (*aliasesTable, error) {
	if _, err := a.db.Exec("insert into aliases (email, alias) values ($1, $2)", email, alias); err != nil {
		return nil, err
	} else {
		return &aliasesTable{Email: email, Alias: alias}, nil
	}
}
