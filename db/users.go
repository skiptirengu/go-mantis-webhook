package db

import (
	"database/sql"
	"errors"
	"github.com/kisielk/sqlstruct"
)

var UserNotFound = errors.New("user not found")

type users struct {
	db *sql.DB
}

type UsersTable struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (u users) Get(email string) (*UsersTable, error) {
	const query = `
		select u.id, u.name, case when u.email is null then a.email else u.email end as email from users u
		left join aliases a on u.email = a.email where u.email = $1 or a.email = $2
	`

	rows, err := u.db.Query(query, email, email)
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, UserNotFound
	}

	res := &UsersTable{}
	sqlstruct.Scan(res, rows)

	return res, nil
}

func (u users) CreateOrUpdate(id int, name, email string) (*UsersTable, error) {
	const query = `
		insert into users (id, name, email) values ($1, $2, $3) on conflict (id) do
		update set (name, email) = ($4, $5) 
	`

	if _, err := u.db.Exec(query, id, name, email, name, email); err != nil {
		return nil, err
	} else {
		return &UsersTable{id, name, email}, nil
	}
}
