package main

import (
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
)

type User struct {
	UserID      int64          `json:"userid"`
	Username    string         `json:"name"`
	Password    string         `json:"password"`
	Email       string         `json:"email"`
	Country     string         `json:"country"`
	Locale      string         `json:"locale"`
	DateCreated mysql.NullTime `json:"date_created"`
	Verified    bool           `json:"verified"`
	Banned      bool           `json:"banned"`
	LastToken   string         `json:"last_token"`
	LastLogin   mysql.NullTime `json:"last_login"`
	LastIP      string         `json:"last_ip"`
}

func (u *User) getUser(db *sql.DB) error {
	return db.QueryRow("SELECT name, email, country, locale, last_token, verified, banned, date_created, last_login FROM users WHERE userid=$1",
		u.UserID).Scan(&u.Username, &u.Email, &u.Country, &u.Locale, &u.LastToken, &u.Verified, &u.Banned, &u.DateCreated, &u.LastLogin)
}

func (u *User) updateUser(db *sql.DB) error {
	_, err := db.Exec("UPDATE users SET name=$1, email=$2, country=$3, locale=$4, FROM users WHERE userid=$5",
		u.Username, u.Email, u.Country, u.Locale, u.UserID)
	return err
}

func (u *User) deleteUser(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM users WHERE userid=$1", u.UserID)
	return err
}

func (u *User) createUser(db *sql.DB) error {
	return db.QueryRow("INSERT INTO users(name, password, email, country, locale) VALUES($1, $2, $3, $4, $5) RETURNING userid",
		u.Username, u.Password, u.Email, u.Country, u.Locale).Scan(&u.UserID)
}

func (u *User) getUsers(db *sql.DB) ([]User, error) {
	return nil, errors.New("Not implimented")
}

func getUsers(db *sql.DB, start, count int) ([]User, error) {

	rows, err := db.Query(
		"SELECT name, email, country, locale, last_token, verified, banned, date_created, last_login FROM users LIMIT $1 OFFSET $2",
		count, start)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []User{}

	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Username, &u.Email, &u.Country, &u.Locale, &u.LastToken, &u.Verified, &u.Banned, &u.DateCreated, &u.LastLogin); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}
