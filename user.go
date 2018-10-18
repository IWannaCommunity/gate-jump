package main

import (
	"database/sql"
	"gate-jump/res"

	"github.com/go-sql-driver/mysql"
)

type User struct {
	ID          int64          `json:"id"`
	Name        string         `json:"name"`
	Password    string         `json:"password"`
	Email       string         `json:"email"`
	Country     string         `json:"country"`
	Locale      string         `json:"locale"`
	DateCreated mysql.NullTime `json:"date_created"`
	Verified    bool           `json:"verified"`
	Banned      bool           `json:"banned"`
	Admin       bool           `json:"admin"`
	LastToken   string         `json:"last_token"`
	LastLogin   mysql.NullTime `json:"last_login"`
	LastIP      string         `json:"last_ip"`
}

func (u *User) getUser(db *sql.DB) *res.ServerError {
	var serr res.ServerError
	serr.Query = "SELECT name, email, country, locale, verified, date_created, last_login FROM users WHERE id=?"
	serr.Args = append(serr.Args, u.ID)
	serr.Err = db.QueryRow(serr.Query, serr.Args...).
		Scan(&u.Name, &u.Email, &u.Country, &u.Locale, &u.Verified, &u.DateCreated, &u.LastLogin)
	return &serr
}

func (u *User) updateUser(db *sql.DB) *res.ServerError {
	var serr res.ServerError
	serr.Query = "UPDATE users SET name=?, email=?, country=?, locale=?, FROM users WHERE id=?"
	serr.Args = append(serr.Args, u.Name, u.Email, u.Country, u.Locale, u.ID)
	_, serr.Err = db.Exec(serr.Query, serr.Args...)
	return &serr
}

func (u *User) deleteUser(db *sql.DB) *res.ServerError {
	var serr res.ServerError
	serr.Query = "DELETE FROM users WHERE id=?"
	serr.Args = append(serr.Args, u.ID)
	_, serr.Err = db.Exec(serr.Query, serr.Args...)
	return &serr
}

func (u *User) createUser(db *sql.DB) *res.ServerError {
	var serr res.ServerError
	var result sql.Result
	serr.Query = "INSERT INTO users(name, password, email, country, locale) VALUES(?, ?, ?, ?, ?)"
	serr.Args = append(serr.Args, u.Name, u.Password, u.Email, u.Country, u.Locale)
	result, serr.Err = db.Exec(serr.Query, serr.Args...)
	if serr.Err != nil {
		return &serr
	}
	u.ID, _ = result.LastInsertId() // we confirmed that there will be no error
	return nil
}

func getUsers(db *sql.DB, start, count int) ([]User, *res.ServerError) {
	var serr res.ServerError
	var rows *sql.Rows
	serr.Query = "SELECT name, email, country, locale, last_token, verified, banned, date_created, last_login FROM users LIMIT ? OFFSET ?"
	serr.Args = append(serr.Args, count, start)
	rows, serr.Err = db.Query(serr.Query, serr.Args...)

	if serr.Err != nil {
		return nil, &serr
	}

	defer rows.Close()

	users := []User{}

	for rows.Next() {
		var u User
		if serr.Err = rows.
			Scan(&u.Name, &u.Email, &u.Country, &u.Locale, &u.LastToken, &u.Verified, &u.Banned, &u.DateCreated, &u.LastLogin); serr.Err != nil {
			return nil, &serr
		}
		users = append(users, u)
	}

	return users, nil
}

func (u *User) getUserByName(db *sql.DB) *res.ServerError {
	var serr res.ServerError
	serr.Query = "SELECT id, email, country, locale, verified, date_created, last_login FROM users WHERE name=?"
	serr.Args = append(serr.Args, u.Name)
	serr.Err = db.QueryRow(serr.Query, serr.Args...).
		Scan(&u.ID, &u.Email, &u.Country, &u.Locale, &u.Verified, &u.DateCreated, &u.LastLogin)
	return &serr
}
