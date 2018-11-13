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
	"time"

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

func request(method string, url string, body interface{}, token interface{}) (int, *Payload, error) {
	var bodyByte []byte

	if body != nil {
		bodyByte = []byte(body.(string))
	}

	req, _ := http.NewRequest(method, url, bytes.NewBuffer(bodyByte))
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
		_, _, _ = request("POST", "/register", newUser, nil)
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
	_, r, _ := request("POST", "/login", fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password), nil)
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

	code, r, err = request(method, route, nil, nil)
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
		code, r, err = request(method, route, badRequest, nil)
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
	code, r, err = request(method, route, mainUser, nil)
	if assert.NoError(t, err) {

		assert.Equal(t, http.StatusCreated, code, "expected statusok")
		assert.True(t, r.Success)

		assert.Nil(t, r.Error)
		assert.Nil(t, r.Token)
		assert.Nil(t, r.User)
		assert.Nil(t, r.UserList)
	}

	// test duplicate username
	code, r, err = request(method, route, duplicateName, nil)
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
	code, r, err = request(method, route, duplicateEmail, nil)
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
		code, r, err = request(method, route, badRequest, nil)
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
	code, r, err = request(method, route, correct, nil)
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
	code, r, err = request(method, route, deleted, nil)
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
	code, r, err = request(method, route, banned, nil)
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
		code, r, err = request(method, route, nil, token)
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

	code, r, err = request(method, route, nil, goodToken)
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
	t.Error("Not Implemented")
}

func TestGetUsers(t *testing.T) {
	t.Error("Not Implemented")
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
