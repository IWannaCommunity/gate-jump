package main

import (
	"database/sql"
	"gate-jump/src/api/res"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type User struct {
	// PUBLIC < USER == ADMINUSER < ADMIN < SERVER
	ID int64 `json:"id"`
	// Read: PUBLIC
	// Write: Nobody
	Name *string `json:"name,omitempty"`
	// Read: PUBLIC
	// Write: USER (offensive names, all included due to offensive names)
	Password *string `json:"password,omitempty"`
	// Read: SERVER
	// Write: USER or SERVER
	Email *string `json:"email,omitempty"`
	// Read: USER
	// Write: USER or SERVER
	Country *string `json:"country,omitempty"`
	// Read: PUBLIC
	// Write: USER or SERVER
	Locale *string `json:"locale,omitempty"`
	// Read: PUBLIC
	// Write: USER or SERVER
	DateCreated *time.Time `json:"date_created,omitempty"`
	// Read: PUBLIC
	// Write: Nobody (This is only ever set on creation.)
	Verified *bool `json:"verified,omitempty"`
	// Read: PUBLIC
	// Write: Nobody (by logging into sql only)
	Banned *bool `json:"banned,omitempty"`
	// Read: PUBLIC
	// Write: ADMIN only
	Admin *bool `json:"admin,omitempty"`
	// Read: PUBLIC
	// Write: Nobody (by logging into sql only)
	LastToken *string `json:"last_token,omitempty"` // ? is this needed
	// Read: SERVER
	// Write: SERVER
	LastLogin *time.Time `json:"last_login,omitempty"`
	// Read: PUBLIC
	// Write: SERVER
	LastIP *string `json:"last_ip,omitempty"`
	// Read: ADMIN
	// Write: SERVER
	Deleted *bool `json:"deleted,omitempty"`
	// READ: ADMIN
	// WRITE: USER (if false) Nobody (if true [sql only])
	DateDeleted *time.Time `json:"date_deleted,omitempty"`
	// READ: ADMIN
	// WRITE: SERVER
}

type UserList struct {
	StartIndex int    `json:"startIndex"` // starting index
	TotalItems int    `json:"totalItems"` // how many items are returned
	Users      []User `json:"users"`      // user array
}

// SQL FUNCTIONS =================================================================================

func (u *User) getUser(db *sql.DB, auth AuthLevel) *res.ServerError {
	var serr res.ServerError
	if auth > USER { // deleted search check
		serr.Query = "SELECT * FROM users WHERE id=?"
	} else {
		serr.Query = "SELECT * FROM users WHERE id=? AND deleted=false"
	}
	serr.Args = append(serr.Args, u.ID)
	serr.Err = u.ScanAll(db.QueryRow(serr.Query, serr.Args...))
	if serr.Err != nil {
		return &serr
	}
	u.CleanDataRead(auth)
	return &serr
}

func (u *User) updateUser(db *sql.DB, auth AuthLevel) *res.ServerError {
	var serr res.ServerError
	//"UPDATE users SET name=?, password=?, email=?, country=?, locale=?, last_token=?, last_login=?, last_ip=? WHERE id=?"
	serr.Query = "UPDATE users SET"

	if u.Name != nil { // USER+
		serr.Query += " name=?,"
		serr.Args = append(serr.Args, u.Name)
	} else { // dont return if not edited
		u.Name = nil
	}
	if u.Password != nil && (auth == USER || auth == ADMINUSER || auth == SERVER) {
		serr.Query += " password=?,"
		serr.Args = append(serr.Args, u.Password) // hash handled in handler
	}
	u.Password = nil // never return password in payload
	if u.Email != nil && (auth == USER || auth == ADMINUSER || auth == SERVER) {
		serr.Query += " email=?,"
		serr.Args = append(serr.Args, u.Email)
	} else {
		u.Email = nil
	}
	if u.Country != nil && (auth == USER || auth == ADMINUSER || auth == SERVER) {
		serr.Query += " country=?,"
		serr.Args = append(serr.Args, u.Country)
	} else {
		u.Country = nil
	}
	if u.Locale != nil && (auth == USER || auth == ADMINUSER || auth == SERVER) {
		serr.Query += " locale=?,"
		serr.Args = append(serr.Args, u.Locale)
	} else {
		u.Locale = nil
	}
	if u.Banned != nil && (auth == ADMIN) {
		serr.Query += " banned=?,"
		serr.Args = append(serr.Args, u.Banned)
	} else {
		u.Banned = nil
	}
	if u.LastToken != nil && (auth == SERVER) {
		serr.Query += " last_token=?,"
		serr.Args = append(serr.Args, u.LastToken)
	} else {
		u.LastToken = nil
	}
	if u.LastLogin != nil && (auth == SERVER) {
		serr.Query += " last_login=?,"
		serr.Args = append(serr.Args, u.LastLogin)
	} else {
		u.LastLogin = nil
	}
	if u.LastIP != nil && (auth == SERVER) {
		serr.Query += " last_ip=?,"
		serr.Args = append(serr.Args, u.LastIP)
	} else {
		u.LastIP = nil
	}

	if len(serr.Args) == 0 {
		return &serr // there were no sections to update so return empty user
	}

	serr.Query = serr.Query[:len(serr.Query)-1] + " WHERE id=?" // remove last comma of query and add WHERE condition
	serr.Args = append(serr.Args, u.ID)

	_, serr.Err = db.Exec(serr.Query, serr.Args...)
	return &serr
}

func (u *User) deleteUser(db *sql.DB) *res.ServerError {
	var serr res.ServerError
	serr.Query = "UPDATE users SET deleted=true AND date_deleted=? WHERE id=?"
	serr.Args = append(serr.Args, time.Now(), u.ID)
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
	serr.Query = "SELECT * FROM users LIMIT ? OFFSET ? WHERE deleted=false"
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
		u.CleanDataRead(auth)
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
		&u.LastIP,
		&u.Deleted,
		&u.DateDeleted)
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
		&u.LastIP,
		&u.Deleted,
		&u.DateDeleted)
}

// applies read user data permissions of a fully retrieved user
func (u *User) CleanDataRead(auth AuthLevel) {
	switch auth {
	case SERVER:
		// we dont want to stop the server from reading anything
	case PUBLIC:
		u.Email = nil
		fallthrough
	case USER:
		u.LastIP = nil
		u.DateDeleted = nil
		u.Deleted = nil
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
		*u.Admin,
		u.Country,
		u.Locale,
		*u.Verified,
		*u.Banned,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(), //expire in one hour
			Issuer:    Config.Host + ":" + Config.Port,
			Subject:   strconv.FormatInt(u.ID, 10), //user id as string
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(Config.JwtSecret))
}
