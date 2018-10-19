package main

import (
	"database/sql"
	"gate-jump/res"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type User struct {
	ID          int64      `json:"id"`
	Name        *string    `json:"name"`
	Password    *string    `json:"password"`
	Email       *string    `json:"email"`
	Country     *string    `json:"country"`
	Locale      *string    `json:"locale"`
	DateCreated *time.Time `json:"date_created"`
	Verified    bool       `json:"verified"`
	Banned      bool       `json:"banned"`
	Admin       bool       `json:"admin"`
	LastToken   *string    `json:"last_token"`
	LastLogin   *time.Time `json:"last_login"`
	LastIP      *string    `json:"last_ip"`
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
	serr.Query = "UPDATE users SET name=?, email=?, country=?, locale=? FROM users WHERE id=?"
	serr.Args = append(serr.Args, u.Name.String, u.Email.String, u.Country.String, u.Locale.String, u.ID)
	_, serr.Err = db.Exec(serr.Query, serr.Args...)
	return &serr
}

func (u *User) updateLoginInfo(db *sql.DB) *res.ServerError {
	var serr res.ServerError
	serr.Query = "UPDATE users SET last_token=?, last_login=?, last_ip=? FROM users WHERE id=?"
	serr.Args = append(serr.Args, u.LastToken, u.LastLogin, u.LastIP, u.ID)
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

// used to determine if valid login username or username in use
func (u *User) getUserByName(db *sql.DB) *res.ServerError {
	var serr res.ServerError
	serr.Query = "SELECT * FROM users WHERE name=?"
	serr.Args = append(serr.Args, u.Name)
	serr.Err = db.QueryRow(serr.Query, serr.Args...).
		Scan(&u.ID, &u.Name, &u.Password, &u.Email, &u.Country, &u.Locale,
			&u.DateCreated, &u.Verified, &u.Banned, &u.Admin,
			&u.LastToken, &u.LastLogin, &u.LastIP)
	return &serr
}

type Claims struct {
	ID       int64   `json:"id"`
	Username *string `json:"username"`
	Admin    bool    `json:"admin"`
	Country  *string `json:"country"`
	Locale   *string `json:"locale"`
	Verified bool    `json:"verified"`
	Banned   bool    `json:"banned"`
	jwt.StandardClaims
}

func (u *User) CreateToken() (string, error) {
	//create and sign the token
	claims := Claims{
		u.ID,
		u.Name,
		u.Admin,
		u.Country,
		u.Locale,
		u.Verified,
		u.Banned,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(), //expire in one hour
			Issuer:    Config.Host + ":" + Config.Port,
			Subject:   strconv.FormatInt(u.ID, 10), //user id as string
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(Config.JwtSecret))
}
