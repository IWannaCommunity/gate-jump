package routers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/IWannaCommunity/gate-jump/src/api/database"
	"github.com/IWannaCommunity/gate-jump/src/api/res"
	"github.com/stretchr/testify/assert"
)

var db *sql.DB

const tableCreationQuery = `CREATE TABLE users (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL,
    password CHAR(60) BINARY NOT NULL,
    email VARCHAR(100),
    country CHAR(2),
    locale VARCHAR(20),
    date_created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    admin BOOL NOT NULL DEFAULT FALSE,
    verified BOOL NOT NULL DEFAULT FALSE,
    banned BOOL NOT NULL DEFAULT FALSE,
    last_token BLOB,
    last_login DATETIME,
    last_ip VARCHAR(50),
    deleted BOOL NOT NULL DEFAULT FALSE,
    date_deleted DATETIME,
    PRIMARY KEY (id)
)`

type Payload struct {
	Success  bool               `json:"success"`
	Error    *res.ResponseError `json:"error,omitempty"`
	Token    *string            `json:"token,omitempty"`
	User     *database.User     `json:"user,omitempty"`
	UserList *database.UserList `json:"userList,omitempty"`
}

func clearTable() {
	db.Exec("DELETE FROM users")
	db.Exec("ALTER TABLE users AUTO_INCREMENT = 1")
}

func ensureTableExists() {
	db.Exec(tableCreationQuery)
}

func createFormValueSet(pairs ...[]string) [][]string {
	var formValueSet [][]string
	for _, pair := range pairs {
		formValueSet = append(formValueSet, pair)
	}
	return formValueSet
}

func createFormValue(key string, value string) []string {
	return []string{key, value}
}

func request(method string, apiUrl string, jsonPayload interface{}, token interface{}, formPayload [][]string) (int, *Payload, error) {
	var reader io.Reader
	form := url.Values{}
	butItWasMeAForm := false

	// init reader
	if jsonPayload != nil {
		reader = bytes.NewBuffer([]byte(jsonPayload.(string)))
	} else if formPayload != nil {
		butItWasMeAForm = true
		for _, pair := range formPayload {
			form.Add(pair[0], pair[1])
		}
		reader = strings.NewReader(form.Encode())
	}

	req, _ := http.NewRequest(method, apiUrl, reader)
	if butItWasMeAForm {
		req.Form = form
	}

	if token != nil {
		req.Header.Set("Authorization", token.(string))
	}
	response := executeRequest(req)
	p, err := unmarshal(response.Body)
	if err != nil {
		return 0, nil, err
	}
	return response.Code, p, nil
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func unmarshal(responseBody *bytes.Buffer) (*Payload, error) {
	var p Payload
	decoder := json.NewDecoder(responseBody)
	if err := decoder.Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

// createUsers creates x amount of users with name:user{i}; password:password{i}; and email:email{i}@website.com
func create(count int) {
	clearTable()
	if count < 1 {
		count = 1
	}

	for i := 1; i < count+1; i++ {
		newUser := fmt.Sprintf(`{"name":"user%d","password":"password%d","email":"email%d@website.com"}`, i, i, i)
		_, _, _ = request("POST", "/register", newUser, nil, nil)
	}
}

func update(id int, country string, locale string, admin bool, banned bool, deleted bool) {
	if deleted {
		db.Exec("UPDATE users SET country=?, locale=?, admin=?, banned=?, deleted=true, date_deleted=? WHERE id=?", country, locale, admin, banned, time.Now(), id)
	} else {
		db.Exec("UPDATE users SET country=?, locale=?, admin=?, banned=?, deleted=false WHERE id=?", country, locale, admin, banned, id)
	}
}

func login(username string, password string) string {
	_, r, _ := request("POST", "/login", fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password), nil, nil)
	if r.Token == nil {
		return ""
	}
	return *r.Token
}

func TestMain(m *testing.M) {
	var err error
	// initialize the database for testing
	//settings.FromFile("config/config.json")
	database.Connect("root", "password", "gatejump")

	// initialize the tests database connection
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8mb4&parseTime=True&interpolateParams=true", "root", "password", "gatejump")) //settings.Database.Username, settings.Database.Password, "usertest"))
	if err != nil {
		log.Panic(err) // can't run tests if we can't initialize the database
	}
	// run router on different thread so we can continue testing
	go Serve("10421", "444")
	// clean and run it
	ensureTableExists()
	clearTable()
	code := m.Run()
	os.Exit(code)
}

func TestAlive(t *testing.T) {
	clearTable()
	var code int
	var r *Payload
	var err error
	method := "GET"
	route := "/"

	code, r, err = request(method, route, nil, nil, nil)
	if assert.NoError(t, err) {

		assert.Equal(t, http.StatusOK, code, "expected statusok")
		assert.True(t, r.Success)

		assert.Nil(t, r.Error)
		assert.Nil(t, r.Token)
		assert.Nil(t, r.User)
		assert.Nil(t, r.UserList)

	}
}

func TestCreateUser(t *testing.T) {
	clearTable()
	var code int
	var r *Payload
	var err error
	method := "POST"
	route := "/register"

	var badRequests []string // only valid request should be one that contains name, email, and password

	badRequests = append(badRequests,
		`sdfdrslkjgnm4momgom!!!`,                                                             // jibberish
		`{"password":"12345678","email":"email@website.com"}`,                                // missing name
		`{"name":"test_user","email":"email@website.com"}`,                                   // missing password
		`{"name":"test_user","password":"12345678"}`,                                         // missing email
		`{"name":"test_user","password":"12345678","country":"us","locale":"en"}`,            // extra
		`{"name":"12356","password":"12345678","email":"email@website.com"}`,                 // invalid username (all numerics)
		`{"name":"test_user@website.com","password":"12345678","email":"email@website.com"}`, // invalid username (its an email)
		`{"name":"test_user","password":"12345","email":"email@website.com"}`,                // invalid password (less than 8 characters)
		`{"name":"test_user","password":"12345678","email":"email"}`)                         // invalid email (non-email format)
	mainUser := `{"name":"test_user","password":"12345678","email":"email@website.com"}`               // valid request
	duplicateName := `{"name":"test_user","password":"12345678","email":"email@someotherwebsite.com"}` // name == mainUser.Name
	duplicateEmail := `{"name":"some_other_user","password":"12345678","email":"email@website.com"}`   // email == mainUser.Email

	// test bad request
	for i, badRequest := range badRequests {
		code, r, err = request(method, route, badRequest, nil, nil)
		if assert.NoError(t, err) {

			assert.Equalf(t, http.StatusBadRequest, code, "excpected badrequest code; badRequest: %d", i)
			assert.Falsef(t, r.Success, "badRequest: %d", i)
			if assert.NotNilf(t, r.Error, "badRequest: %d", i) {
				switch i {
				case 5:
					fallthrough
				case 6: // invalid username
					assert.Equalf(t, "Invalid Username Format", r.Error.Message, "badRequest: %d", i)
				case 7:
					assert.Equalf(t, "Invalid Password Format", r.Error.Message, "badRequest: %d", i)
				case 8:
					assert.Equalf(t, "Invalid Email Format", r.Error.Message, "badRequest: %d", i)
				default:
					assert.Equalf(t, "Invalid Request Payload", r.Error.Message, "badRequest: %d", i)
				}
			}

			assert.Nilf(t, r.Token, "badRequest: %d", i)
			assert.Nilf(t, r.User, "badRequest: %d", i)
			assert.Nilf(t, r.UserList, "badRequest: %d", i)
		}
	}

	// test creating a user
	code, r, err = request(method, route, mainUser, nil, nil)
	if assert.NoError(t, err) {

		assert.Equal(t, http.StatusCreated, code, "expected statusok")
		assert.True(t, r.Success)

		assert.Nil(t, r.Error)
		assert.Nil(t, r.Token)
		assert.Nil(t, r.User)
		assert.Nil(t, r.UserList)
	}

	// test duplicate username
	code, r, err = request(method, route, duplicateName, nil, nil)
	if assert.NoError(t, err) {

		assert.Equal(t, http.StatusConflict, code, "expected statusconflict")
		assert.False(t, r.Success)
		if assert.NotNil(t, r.Error) {
			assert.Equal(t, "Username Already Exists", r.Error.Message)
		}

		assert.Nil(t, r.Token)
		assert.Nil(t, r.User)
		assert.Nil(t, r.UserList)
	}

	// test duplicate email
	code, r, err = request(method, route, duplicateEmail, nil, nil)
	if assert.NoError(t, err) {

		assert.Equal(t, http.StatusConflict, code, "expected statusconflict")
		assert.False(t, r.Success)
		if assert.NotNil(t, r.Error) {
			assert.Equal(t, "Email Already In Use", r.Error.Message)
		}

		assert.Nil(t, r.Token)
		assert.Nil(t, r.User)
		assert.Nil(t, r.UserList)
	}
}

func TestLoginUser(t *testing.T) {
	clearTable()
	create(3)
	update(2, "us", "en", false, false, true) // deleted
	update(3, "us", "en", false, true, false) // banned
	var code int
	var r *Payload
	var err error
	method := "POST"
	route := "/login"

	var badRequests []string // only valid requests are ones that have a correct username and password

	badRequests = append(badRequests,
		`{"BADREQUEST"}`,               // wrong format lol
		`{"password":"wrongpassword"}`, // missing username
		`{"username":"wrongusername"}`, // missing password
		`{"username":"wrongusername","password":"wrongpassword","county":"us"}`, // extra info
		`{"username":"wrongusername","password":"wrongpassword"}`,               // wrong 	username 	wrong password
		`{"username":"wrongusername","password":"password1"}`,                   // wrong 	username 	correct password
		`{"username":"user1","password":"wrongpassword"}`)                       // correct 	username 	wrong password
	correct := `{"username":"user1","password":"password1"}` // correct 	username 	correct password
	deleted := `{"username":"user2","password":"password2"}` // deleted account
	banned := `{"username":"user3","password":"password3"}`  // banned account

	for i, badRequest := range badRequests {
		code, r, err = request(method, route, badRequest, nil, nil)
		if assert.NoError(t, err) {

			switch i {
			case 0:
				fallthrough
			case 1:
				fallthrough
			case 2:
				assert.Equalf(t, http.StatusBadRequest, code, "expected badrequest code; badRequest: %d", i)
			default:
				assert.Equalf(t, http.StatusUnauthorized, code, "expected unauthorized code; badRequest: %d", i)
			}
			assert.Falsef(t, r.Success, "badRequest: %d", i)
			if assert.NotNilf(t, r.Error, "badRequest: %d", i) {
				switch i {
				case 0:
					fallthrough
				case 1:
					fallthrough
				case 2:
					assert.Equalf(t, "Invalid Request Payload", r.Error.Message, "badRequest: %d", i)
				case 3: // extra info, we dont care
					fallthrough
				case 4: // wrong username wrong password
					fallthrough
				case 5: // wrong
					assert.Equalf(t, "User Doesn't Exist", r.Error.Message, "badRequest: %d", i)
				case 6:
					assert.Equalf(t, "Wrong Password", r.Error.Message, "badRequest: %d", i)
				default:
					assert.Truef(t, false, "badRequest: %d", i)
					return // shouldn't occur
				}
			}

			assert.Nilf(t, r.Token, "badRequest: %d", i)
			assert.Nilf(t, r.User, "badRequest: %d", i)
			assert.Nilf(t, r.UserList, "badRequest: %d", i)
		}
	}

	// correct username correct password; not banned; not deleted
	code, r, err = request(method, route, correct, nil, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, code, "expected OK")
		if assert.True(t, r.Success) {
			assert.NotNil(t, r.Token)
		}

		assert.Nil(t, r.Error)
		assert.Nil(t, r.User)
		assert.Nil(t, r.UserList)
	}

	// correct username correct password; not banned; deleted
	code, r, err = request(method, route, deleted, nil, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, code, "expected OK")
		if assert.True(t, r.Success) {
			assert.NotNil(t, r.Token)
		}

		assert.Nil(t, r.Error)
		assert.Nil(t, r.User)
		assert.Nil(t, r.UserList)
	}

	// correct username correct password; banned; not deleted
	code, r, err = request(method, route, banned, nil, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusUnauthorized, code, "expected unauthorized")
		if assert.False(t, r.Success) {
			if assert.NotNil(t, r.Error) {
				assert.Equal(t, "Account Banned", r.Error.Message)
			}
		}

		assert.Nil(t, r.Token)
		assert.Nil(t, r.User)
		assert.Nil(t, r.UserList)
	}
}

func TestRefresh(t *testing.T) {

	clearTable()
	create(1)
	var code int
	var r *Payload
	var err error
	method := "POST"
	route := "/refresh"

	var badRequests []string

	badRequests = append(badRequests,
		"badtoken")

	goodToken := login("user1", "password1")

	for i, token := range badRequests {
		code, r, err = request(method, route, nil, token, nil)
		if assert.NoErrorf(t, err, "badRequest %d", i) {
			assert.Equalf(t, http.StatusUnauthorized, code, "expected unauthorized", "badRequest %d", i)
			assert.Falsef(t, r.Success, "badRequest %d", i)
			if assert.NotNilf(t, r.Error, "badRequest %d", i) {
				assert.Equalf(t, "Invalid Token Provided", r.Error.Message, "badRequest %d", i)
			}

			assert.Nil(t, r.Token, "badRequest %d", i)
			assert.Nil(t, r.User, "badRequest %d", i)
			assert.Nil(t, r.UserList, "badRequest %d", i)
		}
	}

	code, r, err = request(method, route, nil, goodToken, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, code, "expected OK")
		assert.True(t, r.Success)
		assert.NotNil(t, r.Token)

		assert.Nil(t, r.Error)
		assert.Nil(t, r.User)
		assert.Nil(t, r.UserList)
	}
}

func TestGetUser(t *testing.T) {

	clearTable()
	create(10)
	var code int
	var r *Payload
	var err error
	method := "GET"
	route := "/user/" // add id to the end of this

	for i := 1; i <= 8; i++ { // give every user a last-login for tokens and last login date stuff
		_ = login("user"+strconv.Itoa(i), "password"+strconv.Itoa(i))
	}

	update(1, "us", "en", false, false, false) // just a user
	update(2, "us", "en", false, false, true)  // deleted
	update(3, "us", "en", false, true, false)  // banned
	update(4, "us", "en", false, true, true)   // banned and deleted
	update(5, "us", "en", true, false, false)  // admin
	update(6, "us", "en", true, false, true)   // admin and deleted
	update(7, "us", "en", true, true, false)   // admin and banned
	update(8, "us", "en", true, true, true)    // admin and banned and deleted

	update(9, "us", "en", false, false, false) // user perms for login
	update(10, "us", "en", true, false, false) // admin user for login
	userToken := login("user9", "password9")
	adminToken := login("user10", "password10")

	// search for nonexisting user
	code, r, err = request(method, route+"99", nil, nil, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusNotFound, code)
		assert.False(t, r.Success)
		if assert.NotNil(t, r.Error) {
			assert.Equal(t, "User Not Found", r.Error.Message)
		}

		assert.Nil(t, r.Token)
		assert.Nil(t, r.User)
		assert.Nil(t, r.UserList)
	}

	// search self (user)
	code, r, err = request(method, route+"9", nil, userToken, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, code)
		assert.True(t, r.Success)
		if assert.NotNil(t, r.User) {
			assert.NotNil(t, r.User.ID)
			assert.NotNil(t, r.User.Name)
			assert.Nil(t, r.User.Password)
			assert.NotNil(t, r.User.Email)
			assert.NotNil(t, r.User.Country)
			assert.NotNil(t, r.User.Locale)
			assert.NotNil(t, r.User.DateCreated)
			assert.NotNil(t, r.User.Verified)
			assert.NotNil(t, r.User.Banned)
			assert.NotNil(t, r.User.Admin)
			assert.Nil(t, r.User.LastToken)
			assert.NotNil(t, r.User.LastLogin)
			assert.Nil(t, r.User.LastIP)
			assert.Nil(t, r.User.Deleted)
			assert.Nil(t, r.User.DateDeleted)
		}

		assert.Nil(t, r.Error)
		assert.Nil(t, r.Token)
		assert.Nil(t, r.UserList)
	}

	// search self (admin)
	code, r, err = request(method, route+"9", nil, adminToken, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, code)
		assert.True(t, r.Success)
		if assert.NotNil(t, r.User) {
			assert.NotNil(t, r.User.ID)
			assert.NotNil(t, r.User.Name)
			assert.Nil(t, r.User.Password)
			assert.NotNil(t, r.User.Email)
			assert.NotNil(t, r.User.Country)
			assert.NotNil(t, r.User.Locale)
			assert.NotNil(t, r.User.DateCreated)
			assert.NotNil(t, r.User.Verified)
			assert.NotNil(t, r.User.Banned)
			assert.NotNil(t, r.User.Admin)
			assert.Nil(t, r.User.LastToken)
			assert.NotNil(t, r.User.LastLogin)
			assert.NotNil(t, r.User.LastIP)
			assert.NotNil(t, r.User.Deleted)
			assert.Nil(t, r.User.DateDeleted) // because he wasn't deleted dur
		}

		assert.Nil(t, r.Error)
		assert.Nil(t, r.Token)
		assert.Nil(t, r.UserList)
	}
	// search as non-token user
	for id := 1; id <= 8; id++ {
		code, r, err = request(method, route+strconv.Itoa(id), nil, nil, nil)
		if assert.NoError(t, err, "request %d", id) {
			switch id {
			case 1: // regular user
				fallthrough
			case 3: // banned
				fallthrough
			case 5: // admin
				fallthrough
			case 7: // admin and banned
				assert.Equalf(t, http.StatusOK, code, "request %d", id)
				assert.True(t, r.Success, "request %d", id)
				if assert.NotNilf(t, r.User, "request %d", id) {
					assert.NotNilf(t, r.User.ID, "request %d", id)
					assert.NotNilf(t, r.User.Name, "request %d", id)
					assert.Nilf(t, r.User.Password, "request %d", id)
					assert.Nilf(t, r.User.Email, "request %d", id)
					assert.NotNilf(t, r.User.Country, "request %d", id)
					assert.NotNilf(t, r.User.Locale, "request %d", id)
					assert.NotNilf(t, r.User.DateCreated, "request %d", id)
					assert.NotNilf(t, r.User.Verified, "request %d", id)
					assert.NotNilf(t, r.User.Banned, "request %d", id)
					assert.NotNilf(t, r.User.Admin, "request %d", id)
					assert.Nilf(t, r.User.LastToken, "request %d", id)
					assert.NotNilf(t, r.User.LastLogin, "request %d", id)
					assert.Nilf(t, r.User.LastIP, "request %d", id)
					assert.Nilf(t, r.User.Deleted, "request %d", id)
					assert.Nilf(t, r.User.DateDeleted, "request %d", id)
				}
				assert.Nilf(t, r.Error, "request %d", id)
			case 2: // deleted
				fallthrough
			case 4: // banned and deleted
				fallthrough
			case 6: // admin and deleted
				fallthrough
			case 8: // admin and banned and deleted
				assert.Equalf(t, http.StatusNotFound, code, "request %d", id)
				assert.Falsef(t, r.Success, "request %d", id)
				if assert.NotNilf(t, r.Error, "request %d", id) {
					assert.Equal(t, "User Not Found", r.Error.Message, "request %d", id)
				}
				assert.Nilf(t, r.User, "request %d", id)
			}
			assert.Nilf(t, r.Token, "request %d", id)
			assert.Nilf(t, r.UserList, "request %d", id)
		}

	}

	// search as public user
	for id := 1; id <= 8; id++ {
		code, r, err = request(method, route+strconv.Itoa(id), nil, userToken, nil)
		if assert.NoError(t, err, "request %d", id) {
			switch id {
			case 1: // regular user
				fallthrough
			case 3: // banned
				fallthrough
			case 5: // admin
				fallthrough
			case 7: // admin and banned

				assert.Equalf(t, http.StatusOK, code, "request %d", id)
				assert.True(t, r.Success, "request %d", id)
				if assert.NotNilf(t, r.User, "request %d", id) {
					assert.NotNilf(t, r.User.ID, "request %d", id)
					assert.NotNilf(t, r.User.Name, "request %d", id)
					assert.Nilf(t, r.User.Password, "request %d", id)
					assert.Nilf(t, r.User.Email, "request %d", id)
					assert.NotNilf(t, r.User.Country, "request %d", id)
					assert.NotNilf(t, r.User.Locale, "request %d", id)
					assert.NotNilf(t, r.User.DateCreated, "request %d", id)
					assert.NotNilf(t, r.User.Verified, "request %d", id)
					assert.NotNilf(t, r.User.Banned, "request %d", id)
					assert.NotNilf(t, r.User.Admin, "request %d", id)
					assert.Nilf(t, r.User.LastToken, "request %d", id)
					assert.NotNilf(t, r.User.LastLogin, "request %d", id)
					assert.Nilf(t, r.User.LastIP, "request %d", id)
					assert.Nilf(t, r.User.Deleted, "request %d", id)
					assert.Nilf(t, r.User.DateDeleted, "request %d", id)
				}
				assert.Nilf(t, r.Error, "request %d", id)
			default: // deleted users
				assert.Equalf(t, http.StatusNotFound, code, "request %d", id)
				assert.Falsef(t, r.Success, "request %d", id)
				if assert.NotNilf(t, r.Error, "request %d", id) {
					assert.Equal(t, "User Not Found", r.Error.Message, "request %d", id)
				}
				assert.Nilf(t, r.User, "request %d", id)
			}
			assert.Nilf(t, r.Token, "request %d", id)
			assert.Nilf(t, r.UserList, "request %d", id)
		}

		for id := 1; id <= 8; id++ {
			code, r, err = request(method, route+strconv.Itoa(id), nil, adminToken, nil)
			if assert.NoError(t, err, "request %d", id) {
				assert.Equalf(t, http.StatusOK, code, "request %d", id)
				assert.True(t, r.Success, "request %d", id)
				if assert.NotNilf(t, r.User, "request %d", id) {
					assert.NotNilf(t, r.User.ID, "request %d", id)
					assert.NotNilf(t, r.User.Name, "request %d", id)
					assert.Nilf(t, r.User.Password, "request %d", id)
					assert.NotNilf(t, r.User.Email, "request %d", id)
					assert.NotNilf(t, r.User.Country, "request %d", id)
					assert.NotNilf(t, r.User.Locale, "request %d", id)
					assert.NotNilf(t, r.User.DateCreated, "request %d", id)
					assert.NotNilf(t, r.User.Verified, "request %d", id)
					assert.NotNilf(t, r.User.Banned, "request %d", id)
					assert.NotNilf(t, r.User.Admin, "request %d", id)
					assert.Nilf(t, r.User.LastToken, "request %d", id)
					assert.NotNilf(t, r.User.LastLogin, "request %d", id)
					assert.NotNilf(t, r.User.LastIP, "request %d", id)
					assert.NotNilf(t, r.User.Deleted, "request %d", id)
					switch id {
					case 2: // deleted
						fallthrough
					case 4: // banned and deleted
						fallthrough
					case 6: // admin and deleted
						fallthrough
					case 8: // admin and banned and deleted
						assert.NotNilf(t, r.User.DateDeleted, "request %d", id)
					default:
						assert.Nilf(t, r.User.DateDeleted, "request %d", id)
					}
				}

				assert.Nilf(t, r.Error, "request %d", id)
				assert.Nilf(t, r.Token, "request %d", id)
				assert.Nilf(t, r.UserList, "request %d", id)
			}
		}

	}

}

func TestGetUsers(t *testing.T) {
	clearTable()

	var code int
	var r *Payload
	var err error
	method := "GET"
	route := "/user" // add id to the end of this

	var offsetList []string
	var limitList []string
	//var payloads []string
	var formValues [][][]string
	var start int
	var count int

	maxCreated := 75 // max amount of users created during testing
	defaultFormValue := [][]string{createFormValue("start", "0"), createFormValue("count", "10")}

	offsetList = append(offsetList, "-1", "0", "100") // -1: error; 0: good; 100: error?;
	limitList = append(limitList, "-1", "0", "100")   // -1: ret1; 0: ret1; 100: ret50;
	for _, offset := range offsetList {
		for _, limit := range limitList {
			formValues = append(formValues,
				[][]string{createFormValue("start", offset), createFormValue("count", limit)},
			)
		}
	}

	// test empty table
	code, r, err = request(method, route, nil, nil, defaultFormValue)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, code)
		assert.True(t, r.Success)
		if assert.NotNil(t, r.UserList) {
			assert.Equal(t, 0, r.UserList.StartIndex)
			assert.Equal(t, 0, r.UserList.TotalItems)
			assert.Nil(t, r.UserList.Users)
		}

		assert.Nil(t, r.Token)
		assert.Nil(t, r.User)
		assert.Nil(t, r.Error)
	}

	// test bad request
	code, r, err = request(method, route, nil, nil, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, code)
		assert.True(t, r.Success)
		if assert.NotNil(t, r.UserList) {
			assert.Equal(t, 0, r.UserList.StartIndex)
			assert.Equal(t, 0, r.UserList.TotalItems)
			assert.Nil(t, r.UserList.Users)
		}

		assert.Nil(t, r.Token)
		assert.Nil(t, r.User)
		assert.Nil(t, r.Error)
	}

	create(maxCreated)

	// test all offsets
	for i, form := range formValues {
		startGiven, _ := strconv.Atoi(formValues[i][0][1]) // look at first form value in set i
		countGiven, _ := strconv.Atoi(formValues[i][1][1]) // look at second form value in set i

		// determine what getUsers should be using
		if startGiven < 0 {
			start = 0
		} else {
			start = startGiven
		}
		if countGiven < 0 {
			count = 0
		} else if countGiven > 50 {
			count = 50
		} else {
			count = countGiven
		}

		// find out how many users should actually get returned
		if count > maxCreated-start {
			count = maxCreated - start
			if count < 0 {
				count = 0
			}
		}

		code, r, err = request(method, route, nil, nil, form)
		assert.Equalf(t, http.StatusOK, code, "{start: %d; count: %d}", startGiven, countGiven)
		assert.Truef(t, r.Success, "{start: %d; count: %d}", startGiven, countGiven)
		if assert.NotNilf(t, r.UserList, "{start: %d; count: %d}", startGiven, countGiven) {
			assert.Equalf(t, start, r.UserList.StartIndex, "{start: %d; count: %d}", startGiven, countGiven)
			assert.Equalf(t, count, r.UserList.TotalItems, "{start: %d; count: %d}", startGiven, countGiven)
			if count > 0 {
				assert.NotNilf(t, r.UserList.Users, "{start: %d; count: %d}", startGiven, countGiven)
			} else {
				assert.Nilf(t, r.UserList.Users, "{start: %d; count: %d}", startGiven, countGiven)
			}

			assert.Nilf(t, r.Token, "{start: %d; count: %d}", startGiven, countGiven)
			assert.Nilf(t, r.User, "{start: %d; count: %d}", startGiven, countGiven)
			assert.Nilf(t, r.Error, "{start: %d; count: %d}", startGiven, countGiven)
		}
	}
}

func TestGetUserByName(t *testing.T) {
	t.Error("Not Implemented")
}

func TestUpdateUser(t *testing.T) {
	t.Error("Not Implemented")
}

func TestDeleteUser(t *testing.T) {
	t.Error("Not Implemented")
}
