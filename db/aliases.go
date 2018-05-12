package db

import "database/sql"

var Aliases = aliases{Get()}

type aliases struct {
	db *sql.DB
}

type aliasesTable struct {
	ID    int    `json:"id"`
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
	var insertedId int
	if res, err := a.db.Query("insert into aliases (email, alias) values ($1, $2) returning id", email, alias); err != nil {
		return nil, err
	} else {
		ScanCol(res, &insertedId)
		return &aliasesTable{ID: insertedId, Email: email, Alias: alias}, nil
	}
}
