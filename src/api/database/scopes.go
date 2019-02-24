package database

import (
	"database/sql"

	"github.com/IWannaCommunity/gate-jump/src/api/res"
)

type Scope struct {
	ID          int64   `json:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func CreateScope(name, description string) (res.ServerError, *Scope) {
	s := &Scope{
		Name:        &name,
		Description: &description,
	}

	serr := *new(res.ServerError)
	serr.Query = "INSERT INTO scopes(name, description) VALUES(?, ?)"
	serr.Args = append(serr.Args, name, description)

	result := *new(sql.Result)
	result, serr.Err = db.Exec(serr.Query, serr.Args...)
	if serr.Err != nil {
		return serr, s
	}
	s.ID, _ = result.LastInsertId()
	return serr, s
}
