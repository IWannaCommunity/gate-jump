package main

import (
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
)

type User struct {
	UserID      int64          `json:"id"`
	Username    string         `json:"name"`
	Password    string         `json:"password"`
	Email       string         `json:"email"`
	Country     string         `json:"country"`
	Language    string         `json:"lang"`
	LastToken   string         `json:"last_token"`
	LastIP      string         `json:"last_ip"`
	Verified    bool           `json:"verified"`
	Banned      bool           `json:"banned"`
	DateCreated mysql.NullTime `json:"date_created"`
	LastLogin   mysql.NullTime `json:"last_login"`
}

func (u *User) getUser(db *sql.DB) error {
	return errors.New("Not implimented")
}

func (u *User) updateUser(db *sql.DB) error {
	return errors.New("Not implimented")
}

func (u *User) deleteUser(db *sql.DB) error {
	return errors.New("Not implimented")
}

func (u *User) createUser(db *sql.DB) error {
	return errors.New("Not implimented")
}

func (u *User) getUsers(db *sql.DB) ([]User, error) {
	return nil, errors.New("Not implimented")
}
