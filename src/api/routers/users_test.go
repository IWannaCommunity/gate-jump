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

	// test invalid username

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
