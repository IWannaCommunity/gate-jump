package routers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/IWannaCommunity/gate-jump/src/api/database"
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
	Success  bool          `json:"success"`
	Error    *PayloadError `json:"error,omitempty"`
	Token    *string       `json:"token,omitempty"`
	User     interface{}   `json:"user,omitempty"`
	UserList interface{}   `json:"userList,omitempty"`
}

type PayloadError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func clearTable() {
	db.Exec("DELETE FROM users")
	db.Exec("ALTER TABLE users AUTO_INCREMENT = 1")
}

func ensureTableExists() {
	db.Exec(tableCreationQuery)
}

func request(method string, url string, body []byte) (int, *Payload, error) {
	req, _ := http.NewRequest(method, url, bytes.NewBuffer(body))
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
		newUser := []byte(fmt.Sprintf(`{"name":"user%d","password":"password%d","email":"email%d@website.com"}`, i, i, i))
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(newUser))
		_ = executeRequest(req)
	}
}

func update(id int, country string, locale string, admin bool, banned bool, deleted bool) {
	if deleted {
		db.Exec("UPDATE users SET country=?, locale=?, admin=?, banned=?, deleted=true, date_deleted=? WHERE id=?", country, locale, admin, banned, time.Now(), id)
	} else {
		db.Exec("UPDATE users SET country=?, locale=?, admin=?, banned=?, deleted=false WHERE id=?", country, locale, admin, banned, id)
	}
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

	code, r, err = request(method, route, nil)
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

	var badRequests [8][]byte // only valid request should be one that contains name, email, and password

	badRequests[0] = []byte(`sdfdrslkjgnm4momgom!!!`)                                                          // jibberish
	badRequests[1] = []byte(`{"password":"12345678","email":"email@website.com"}`)                             // missing name
	badRequests[2] = []byte(`{"name":"test_user","email":"email@website.com"}`)                                // missing password
	badRequests[3] = []byte(`{"name":"test_user","password":"12345678"}`)                                      // missing email
	badRequests[4] = []byte(`{"name":"test_user","password":"12345678","country":"us","locale":"en"}`)         // extra
	badRequests[5] = []byte(`{"name":"12356","password":"12345678","email":"email@website.com"}`)              // invalid username (all numerics)
	badRequests[6] = []byte(`{"name":"test_user","password":"12345","email":"email@website.com"}`)             // invalid password (less than 8 characters)
	badRequests[7] = []byte(`{"name":"test_user","password":"12345678","email":"email"}`)                      // invalid email (non-email format)
	mainUser := []byte(`{"name":"test_user","password":"12345678","email":"email@website.com"}`)               // valid request
	duplicateName := []byte(`{"name":"test_user","password":"12345678","email":"email@someotherwebsite.com"}`) // name == mainUser.Name
	duplicateEmail := []byte(`{"name":"some_other_user","password":"12345678","email":"email@website.com"}`)   // email == mainUser.Email

	// test bad request
	for i, badRequest := range badRequests {
		code, r, err = request(method, route, badRequest)
		if assert.NoError(t, err) {

			assert.Equalf(t, http.StatusBadRequest, code, "excpected badrequest code; badRequest: %d", i)
			assert.Falsef(t, r.Success, "badRequest: %d", i)
			if assert.NotNilf(t, r.Error, "badRequest: %d", i) {
				switch i {
				case 5: // invalid username
					assert.Equalf(t, "Invalid Username Format", r.Error.Message, "badRequest: %d", i)
				case 6:
					assert.Equalf(t, "Invalid Password Format", r.Error.Message, "badRequest: %d", i)
				case 7:
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
	code, r, err = request(method, route, mainUser)
	if assert.NoError(t, err) {

		assert.Equal(t, http.StatusCreated, code, "expected statusok")
		assert.True(t, r.Success)

		assert.Nil(t, r.Error)
		assert.Nil(t, r.Token)
		assert.Nil(t, r.User)
		assert.Nil(t, r.UserList)
	}

	// test duplicate username
	code, r, err = request(method, route, duplicateName)
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
	code, r, err = request(method, route, duplicateEmail)
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

	var badRequests [7][]byte // only valid requests are ones that have a correct username and password

	badRequests[0] = []byte(`{"BADREQUEST"}`)                                                        // wrong format lol
	badRequests[1] = []byte(`{"password":"wrongpassword"}`)                                          // missing username
	badRequests[2] = []byte(`{"username":"wrongusername"}`)                                          // missing password
	badRequests[3] = []byte(`{"username":"wrongusername","password":"wrongpassword","county":"us"}`) // extra info
	badRequests[4] = []byte(`{"username":"wrongusername","password":"wrongpassword"}`)               // wrong 	username 	wrong password
	badRequests[5] = []byte(`{"username":"wrongusername","password":"password1"}`)                   // wrong 	username 	correct password
	badRequests[6] = []byte(`{"username":"user1","password":"wrongpassword"}`)                       // correct 	username 	wrong password
	correct := []byte(`{"username":"user1","password":"password1"}`)                                 // correct 	username 	correct password
	deleted := []byte(`{"username":"user2","password":"password2"}`)                                 // deleted account
	banned := []byte(`{"username":"user3","password":"password3"}`)                                  // banned account

	for i, badRequest := range badRequests {
		code, r, err = request(method, route, badRequest)
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
	code, r, err = request(method, route, correct)
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
	code, r, err = request(method, route, deleted)
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
	code, r, err = request(method, route, banned)
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

func TestGetUser(t *testing.T) {
	/*
		clearTable()
		var code int
		var r *Payload
		var err error
		method := "GET"
		route := "/user/" // add # to any request
	*/
}
