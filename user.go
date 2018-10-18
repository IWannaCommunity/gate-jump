package main

import (
	"database/sql"

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

func (u *User) getUser(db *sql.DB) error {
	return db.QueryRow("SELECT name, email, country, locale, verified, date_created, last_login FROM users WHERE userid=?",
		u.UserID).Scan(&u.Username, &u.Email, &u.Country, &u.Locale, &u.Verified, &u.DateCreated, &u.LastLogin)
}

func (u *User) updateUser(db *sql.DB) error {
	_, err := db.Exec("UPDATE users SET name=?, email=?, country=?, locale=?, FROM users WHERE userid=?",
		u.Username, u.Email, u.Country, u.Locale, u.UserID)
	return err
}

func (u *User) deleteUser(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM users WHERE userid=?", u.UserID)
	return err
}

func (u *User) createUser(db *sql.DB) error {
	result, err := db.Exec("INSERT INTO users(name, password, email, country, locale) VALUES(?, ?, ?, ?, ?)",
		u.Username, u.Password, u.Email, u.Country, u.Locale)
	if err == nil {
		u.UserID, err = result.LastInsertId()
	}
	return err
}

func getUsers(db *sql.DB, start, count int) ([]User, error) {
	rows, err := db.Query(
		"SELECT name, email, country, locale, last_token, verified, banned, date_created, last_login FROM users LIMIT ? OFFSET ?",
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

func (u *User) getUserByName(db *sql.DB) error {
	return db.QueryRow("SELECT userid, email, country, locale, verified, date_created, last_login FROM users WHERE name=?",
		u.Username).Scan(&u.UserID, &u.Email, &u.Country, &u.Locale, &u.Verified, &u.DateCreated, &u.LastLogin)
}
