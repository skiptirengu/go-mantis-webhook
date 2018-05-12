package db

import (
	"database/sql"
)

var Aliases = aliases{Get()}

type aliases struct {
	db *sql.DB
}

type aliasesTable struct {
	Email string `json:"email"`
	Alias string `json:"alias"`
}

func (a aliases) CheckExist(alias string) (bool) {
	var count int
	res, _ := a.db.Query("select count(*) from aliases where alias = $1", alias)
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
