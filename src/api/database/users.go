package database

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/IWannaCommunity/gate-jump/src/api/authentication"
	"github.com/IWannaCommunity/gate-jump/src/api/res"
	"github.com/IWannaCommunity/gate-jump/src/api/settings"
	jwt "github.com/dgrijalva/jwt-go"
)

type User struct {
	// PUBLIC < USER == ADMINUSER < ADMIN < SERVER
	ID *int64 `json:"id",omitempty`
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
	UUID *string `json:"uuid"`
	// READ: PUBLIC
	// WRITE: Nobody
}

type UserList struct {
	StartIndex int    `json:"startIndex"`      // starting index
	TotalItems int    `json:"totalItems"`      // how many items are returned
	Users      []User `json:"users,omitempty"` // user array
}

// SQL FUNCTIONS =================================================================================

func (u *User) GetUser(auth authentication.Level) res.ServerError {
	var serr res.ServerError
	if auth > authentication.USER { // deleted search check
		serr.Query = "SELECT * FROM users WHERE id=?"
	} else {
		serr.Query = "SELECT * FROM users WHERE id=? AND deleted=FALSE"
	}
	serr.Args = append(serr.Args, u.ID)
	serr.Err = u.ScanAll(db.QueryRow(serr.Query, serr.Args...))
	if serr.Err != nil {
		return serr
	}
	u.CleanDataRead(auth)
	return serr
}

func (u *User) GetUserByName(auth authentication.Level) res.ServerError {
	var serr res.ServerError
	if auth > authentication.USER { // deleted search check
		serr.Query = "SELECT * FROM users WHERE name=?"
	} else {
		serr.Query = "SELECT * FROM users WHERE name=? AND deleted=FALSE"
	}
	serr.Args = append(serr.Args, u.Name)
	serr.Err = u.ScanAll(db.QueryRow(serr.Query, serr.Args...))
	if serr.Err != nil {
		return serr
	}
	u.CleanDataRead(auth)
	return serr
}

func (u *User) GetUserByEmail(auth authentication.Level) res.ServerError {
	var serr res.ServerError
	serr.Query = "SELECT * FROM users WHERE email=?"
	serr.Args = append(serr.Args, u.Email)
	serr.Err = u.ScanAll(db.QueryRow(serr.Query, serr.Args...))
	if serr.Err != nil {
		return serr
	}
	u.CleanDataRead(auth)
	return serr
}

func (u *User) UpdateUser(auth authentication.Level) res.ServerError {
	var serr res.ServerError
	//"UPDATE users SET name=?, password=?, email=?, country=?, locale=?, last_token=?, last_login=?, last_ip=? WHERE id=?"
	serr.Query = "UPDATE users SET"

	if u.Name != nil { // USER+
		serr.Query += " name=?,"
		serr.Args = append(serr.Args, u.Name)
	} else { // dont return if not edited
		u.Name = nil
	}
	if u.Password != nil && (auth == authentication.USER || auth == authentication.ADMINUSER || auth == authentication.SERVER) {
		serr.Query += " password=?,"
		serr.Args = append(serr.Args, u.Password) // hash handled in handler
	}
	u.Password = nil // never return password in payload
	if u.Email != nil && (auth == authentication.USER || auth == authentication.ADMINUSER || auth == authentication.SERVER) {
		serr.Query += " email=?,"
		serr.Args = append(serr.Args, u.Email)
	} else {
		u.Email = nil
	}
	if u.Country != nil && (auth == authentication.USER || auth == authentication.ADMINUSER || auth == authentication.SERVER) {
		serr.Query += " country=?,"
		serr.Args = append(serr.Args, u.Country)
	} else {
		u.Country = nil
	}
	if u.Locale != nil && (auth == authentication.USER || auth == authentication.ADMINUSER || auth == authentication.SERVER) {
		serr.Query += " locale=?,"
		serr.Args = append(serr.Args, u.Locale)
	} else {
		u.Locale = nil
	}
	if u.Banned != nil && (auth == authentication.ADMIN) {
		serr.Query += " banned=?,"
		serr.Args = append(serr.Args, u.Banned)
	} else {
		u.Banned = nil
	}
	if u.LastToken != nil && (auth == authentication.SERVER) {
		serr.Query += " last_token=?,"
		serr.Args = append(serr.Args, u.LastToken)
	} else {
		u.LastToken = nil
	}
	if u.LastLogin != nil && (auth == authentication.SERVER) {
		serr.Query += " last_login=?,"
		serr.Args = append(serr.Args, u.LastLogin)
	} else {
		u.LastLogin = nil
	}
	if u.LastIP != nil && (auth == authentication.SERVER) {
		serr.Query += " last_ip=?,"
		serr.Args = append(serr.Args, u.LastIP)
	} else {
		u.LastIP = nil
	}
	if u.Verified != nil && (auth == authentication.SERVER) {
		serr.Query += " verified=?,"
		serr.Args = append(serr.Args, u.Verified)
	} else {
		u.Verified = nil
	}

	if len(serr.Args) == 0 {
		return serr // there were no sections to update so return empty user
	}

	serr.Query = serr.Query[:len(serr.Query)-1] + " WHERE id=?" // remove last comma of query and add WHERE condition
	serr.Args = append(serr.Args, u.ID)

	_, serr.Err = db.Exec(serr.Query, serr.Args...)
	return serr
}

func (u *User) DeleteUser() res.ServerError {
	var serr res.ServerError
	serr.Query = "UPDATE users SET deleted=TRUE, date_deleted=? WHERE id=?"
	serr.Args = append(serr.Args, time.Now(), u.ID)
	_, serr.Err = db.Exec(serr.Query, serr.Args...)
	return serr
}

func (u *User) CreateUser() res.ServerError {
	var serr res.ServerError
	var result sql.Result
	serr.Query = "INSERT INTO users(name, password, email) VALUES(?, ?, ?)"
	serr.Args = append(serr.Args, u.Name, u.Password, u.Email)
	result, serr.Err = db.Exec(serr.Query, serr.Args...)
	if serr.Err != nil {
		return serr
	}
	id, _ := result.LastInsertId() // we confirmed that there will be no error
	u.ID = &id
	return serr
}

func GetUsers(start, count int, auth authentication.Level) (*UserList, res.ServerError) {
	var serr res.ServerError
	var rows *sql.Rows
	serr.Query = "SELECT * FROM users  WHERE deleted=FALSE LIMIT ? OFFSET ?"
	serr.Args = append(serr.Args, count, start)
	rows, serr.Err = db.Query(serr.Query, serr.Args...)

	if serr.Err != nil {
		return nil, serr
	}

	defer rows.Close()

	users := []User{}

	for rows.Next() {
		var u User
		if serr.Err = u.ScanAlls(rows); serr.Err != nil {
			return nil, serr
		}
		u.CleanDataRead(auth)
		users = append(users, u)
	}

	return &UserList{Users: users, StartIndex: start, TotalItems: len(users)}, serr
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
		&u.Admin,
		&u.Verified,
		&u.Banned,
		&u.LastToken,
		&u.LastLogin,
		&u.LastIP,
		&u.Deleted,
		&u.DateDeleted,
		&u.UUID)
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
		&u.Admin,
		&u.Verified,
		&u.Banned,
		&u.LastToken,
		&u.LastLogin,
		&u.LastIP,
		&u.Deleted,
		&u.DateDeleted,
		&u.UUID)
}

// applies read user data permissions of a fully retrieved user
func (u *User) CleanDataRead(auth authentication.Level) {
	switch auth {
	case authentication.SERVER:
		// we dont want to stop the server from reading anything
	case authentication.PUBLIC:
		u.Email = nil
		fallthrough
	case authentication.USER:
		u.LastIP = nil
		u.DateDeleted = nil
		u.Deleted = nil
		fallthrough
	case authentication.ADMINUSER:
		fallthrough
	case authentication.ADMIN:
		fallthrough
	default: // by default always remove password. this is here for security of passwords
		u.Password = nil
		u.LastToken = nil
	}
}

func (u *User) CreateToken() (string, error) {
	//create and sign the token
	claims := authentication.Claims{
		u.UUID,
		u.Name,
		u.Country,
		u.Locale,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(), //expire in one hour
			Issuer:    settings.Host + ":" + settings.Port,
			Subject:   *u.UUID, //user id as string
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(settings.JwtSecret))
}

func (u *User) UnflagDeletion() res.ServerError {
	var serr res.ServerError
	serr.Query = "UPDATE users SET deleted=FALSE, date_deleted=NULL WHERE id=?"
	serr.Args = append(serr.Args, u.ID)
	_, serr.Err = db.Exec(serr.Query, serr.Args...)
	return serr
}

// literally for just debugging
func (u *User) ToString() string {
	var str string
	tabAmount := " ------ "
	str += "\n{"
	str += "\tID:" + tabAmount + strconv.FormatInt(*u.ID, 10) + "\n"
	if u.Name != nil {
		str += "\tName:" + tabAmount + *u.Name + "\n"
	}
	if u.Password != nil {
		str += "\tPassword:" + tabAmount + "notnil" + "\n"
	}
	if u.Email != nil {
		str += "\tEmail:" + tabAmount + *u.Email + "\n"
	}
	if u.Country != nil {
		str += "\tCountry:" + tabAmount + *u.Country + "\n"
	}
	if u.Locale != nil {
		str += "\tLocale:" + tabAmount + *u.Locale + "\n"
	}
	if u.DateCreated != nil {
		str += "\tDateCreated:" + tabAmount + (*u.DateCreated).String() + "\n"
	}
	if u.Admin != nil {
		str += "\tAdmin:" + tabAmount + strconv.FormatBool(*u.Admin) + "\n"
	}
	if u.Verified != nil {
		str += "\tVerified:" + tabAmount + strconv.FormatBool(*u.Verified) + "\n"
	}
	if u.Banned != nil {
		str += "\tBanned:" + tabAmount + strconv.FormatBool(*u.Banned) + "\n"
	}
	if u.LastToken != nil {
		str += "\tLastToken:" + tabAmount + "notnil" + "\n"
	}
	if u.LastLogin != nil {
		str += "\tLastLogin:" + tabAmount + (*u.LastLogin).String() + "\n"
	}
	if u.LastIP != nil {
		str += "\tLastIP:" + tabAmount + *u.LastIP + "\n"
	}
	if u.Deleted != nil {
		str += "\tDeleted:" + tabAmount + strconv.FormatBool(*u.Deleted) + "\n"
	}
	if u.DateDeleted != nil {
		str += "\tDateDeleted:" + tabAmount + (*u.DateDeleted).String() + "\n"
	}
	str += "}"
	return str
}
