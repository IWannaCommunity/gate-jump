package main

import (
	"database/sql"
	"errors"
	"gate-jump/res"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type User struct {
	// PUBLIC < USER == ADMINUSER < ADMIN < SERVER
	ID int64 `json:"id"`
	// Read: PUBLIC
	// Write: Nobody
	Name *string `json:"name"`
	// Read: PUBLIC
	// Write: USER (offensive names, all included due to offensive names)
	Password *string `json:"password,omitempty"`
	// Read: SERVER
	// Write: USER or SERVER
	Email *string `json:"email,omitempty"`
	// Read: USER
	// Write: USER or SERVER
	Country *string `json:"country"`
	// Read: PUBLIC
	// Write: USER or SERVER
	Locale *string `json:"locale"`
	// Read: PUBLIC
	// Write: USER or SERVER
	DateCreated *time.Time `json:"date_created"`
	// Read: PUBLIC
	// Write: Nobody (This is only ever set on creation.)
	Verified bool `json:"verified"`
	// Read: PUBLIC
	// Write: Nobody (by logging into sql only)
	Banned bool `json:"banned"`
	// Read: PUBLIC
	// Write: ADMIN only
	Admin bool `json:"admin"`
	// Read: PUBLIC
	// Write: Nobody (by logging into sql only)
	LastToken *string `json:"last_token,omitempty"` // ? is this needed
	// Read: SERVER
	// Write: SERVER
	LastLogin *time.Time `json:"last_login"`
	// Read: PUBLIC
	// Write: SERVER
	LastIP *string `json:"last_ip,omitempty"`
	// Read: ADMIN
	// Write: SERVER
}

type UserList struct {
	StartIndex int    `json:"startIndex"` // starting index
	TotalItems int    `json:"totalItems"` // how many items are returned
	Users      []User `json:"users"`      // user array
}

// SQL FUNCTIONS =================================================================================

func (u *User) getUser(db *sql.DB, auth AuthLevel) *res.ServerError {
	var serr res.ServerError
	serr.Query = "SELECT * FROM users WHERE id=?"
	serr.Args = append(serr.Args, u.ID)
	serr.Err = u.ScanAll(db.QueryRow(serr.Query, serr.Args...))
	u.CleanDataRead(auth, serr)
	return &serr
}

func (u *User) updateUser(db *sql.DB, auth AuthLevel) *res.ServerError {
	var serr res.ServerError
	switch auth {
	case SERVER: // name, password, email, country, locale, last_token, last_login, last_ip
		serr.Query = "UPDATE users SET name=?, password=?, email=?, country=?, locale=?, last_token=?, last_login=?, last_ip=? WHERE id=?"
		serr.Args = append(serr.Args, u.Name, u.Password, u.Email, u.Country, u.Locale, u.LastToken, u.LastLogin, u.LastIP, u.ID)
	case ADMIN: // can update: name, banned
		serr.Query = "UPDATE users SET name=?, banned=? WHERE id=?"
		serr.Args = append(serr.Args, u.Name, u.Banned, u.ID)
	case ADMINUSER: // can update: name, password, email, country, locale, (cant update own banned status)
		fallthrough // same perms as user
	case USER: // can update: name, password, email, country, locale
		serr.Query = "UPDATE users SET name=?, password=?, email=?, country=?, locale=? WHERE id=?"
		serr.Args = append(serr.Args, u.Name, u.Password, u.Email, u.Country, u.Locale, u.ID)
	case PUBLIC: // can update: nothing (shouldn't occur)
		fallthrough
	default: // can update: nothing (shouldn't occur)
		return &res.ServerError{Err: errors.New("This shouldn't occur.")}
	}
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
	serr.Query = "INSERT INTO users(name, password, email) VALUES(?, ?, ?)"
	serr.Args = append(serr.Args, u.Name, u.Password, u.Email)
	result, serr.Err = db.Exec(serr.Query, serr.Args...)
	if serr.Err != nil {
		return &serr
	}
	u.ID, _ = result.LastInsertId() // we confirmed that there will be no error
	return nil
}

func getUsers(db *sql.DB, start, count int, auth AuthLevel) (*UserList, *res.ServerError) {
	var serr res.ServerError
	var rows *sql.Rows
	serr.Query = "SELECT * FROM users LIMIT ? OFFSET ?"
	serr.Args = append(serr.Args, count, start)
	rows, serr.Err = db.Query(serr.Query, serr.Args...)

	if serr.Err != nil {
		return nil, &serr
	}

	defer rows.Close()

	users := []User{}

	for rows.Next() {
		var u User
		if serr.Err = u.ScanAlls(rows); serr.Err != nil {
			return nil, &serr
		}
		u.CleanDataRead(auth, serr)
		users = append(users, u)
	}

	return &UserList{Users: users, StartIndex: start, TotalItems: len(users)}, nil
}

// HELPER FUNCTIONS ==============================================================================

// scans all user data into the user struct
func (u *User) ScanAll(row *sql.Row) error {
	return row.Scan(
		&u.ID,
		&u.Name,
		&u.Password,
		&u.Email,
		&u.Country,
		&u.Locale,
		&u.DateCreated,
		&u.Verified,
		&u.Banned,
		&u.Admin,
		&u.LastToken,
		&u.LastLogin,
		&u.LastIP)
}

// scans all user data into the user struct (for rows)
func (u *User) ScanAlls(rows *sql.Rows) error {
	return rows.Scan(
		&u.ID,
		&u.Name,
		&u.Password,
		&u.Email,
		&u.Country,
		&u.Locale,
		&u.DateCreated,
		&u.Verified,
		&u.Banned,
		&u.Admin,
		&u.LastToken,
		&u.LastLogin,
		&u.LastIP)
}

// applies read user data permissions of a fully retrieved user
func (u *User) CleanDataRead(auth AuthLevel, serr res.ServerError) {
	if serr.Err != nil {
		return
	}
	switch auth {
	case SERVER:
		// we dont want to stop the server from reading anything
	case PUBLIC:
		u.Email = nil
		fallthrough
	case USER:
		u.LastIP = nil
		fallthrough
	case ADMINUSER:
		fallthrough
	case ADMIN:
		fallthrough
	default: // by default always remove password. this is here for security of passwords
		u.Password = nil
		u.LastToken = nil
	}
}

// used to determine if valid login username or username in use
func (u *User) GetUserByName(db *sql.DB) *res.ServerError {
	var serr res.ServerError
	serr.Query = "SELECT * FROM users WHERE name=?"
	serr.Args = append(serr.Args, u.Name)
	serr.Err = db.QueryRow(serr.Query, serr.Args...).
		Scan(&u.ID, &u.Name, &u.Password, &u.Email, &u.Country, &u.Locale,
			&u.DateCreated, &u.Verified, &u.Banned, &u.Admin,
			&u.LastToken, &u.LastLogin, &u.LastIP)
	return &serr
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
