package main_test

import (
	"bytes"
	"encoding/json"
	main "gate-jump"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var s main.Server

// unmarshal will create a map from a json response, for quick and dirty testing
func unmarshal(t *testing.T, responseBody *bytes.Buffer) map[string]interface{} {
	var result map[string]interface{}
	if err := json.Unmarshal(responseBody.Bytes(), &result); err != nil {
		t.Error(err)
		return nil
	}
	return result
}

// TestMain is the main entrypoint all tests, and ensures the server is ready to go
// every test will call this first
func TestMain(m *testing.M) {
	//init server
	if err := main.LoadConfig("config/config.json"); err != nil {
		panic(err)
	}
	s = main.Server{}
	s.Initialize(main.Config.Database.Username, main.Config.Database.Password, main.Config.Database.Dsn)
	s.InitializeRoutes()
	ensureTableExists()
	clearTable()

	code := m.Run()

	os.Exit(code)
}

func TestEmptyTable(t *testing.T) {
	req, _ := http.NewRequest("GET", "/user", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	result := unmarshal(t, response.Body)
	if result["success"] != true {
		t.Errorf("Expected 'true' got '%v'", result["success"])
	}
	userList := result["userList"].(map[string]interface{})
	if userList["startIndex"].(float64) != 0.0 {
		t.Errorf("Expected '0' got '%v'", userList["totalItems"].(int))
	}
	if userList["totalItems"].(float64) != 0.0 {
		t.Errorf("Expected '0' got '%v'", userList["totalItems"].(int))
	}
}

func TestGetNonExistentUser(t *testing.T) {
	clearTable()
	req, _ := http.NewRequest("GET", "/user/11", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
	result := unmarshal(t, response.Body)
	if result["success"] != false {
		t.Errorf("Expected 'false' got '%v'", result["success"])
	}
	if val := result["error"].(map[string]interface{})["message"].(string); val != "User Not Found" {
		t.Errorf("Expected 'User Not Found' got '%v'", val)
	}
}

func TestCreateUser(t *testing.T) {
	payload := []byte(`{"name":"test user","password":"12345","email":"email@website.com","country":"us","locale":"en"}`)

	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(payload))
	response := executeRequest(req)

	checkResponseCode(t, http.StatusCreated, response.Code)

	m := unmarshal(t, response.Body)
	if m["name"] != "test user" {
		t.Errorf("Expected user name to be 'test user'. Got '%v'", m["name"])
	}

	if m["email"] != "email@website.com" {
		t.Errorf("Expected user email to be 'email@website.com'. Got '%v'", m["email"])
	}

	if m["id"] != 1.0 { // unmarshal converts int to float
		t.Errorf("Expected user ID to be '1'. Got '%v'", m["id"])
	}
}

func TestGetUser(t *testing.T) {
	addUsers(1)

	req, _ := http.NewRequest("GET", "/user/1", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)
}

func addUsers(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		s.DB.Exec("INSERT INTO users(name, password, email) VALUES(?, ?, ?)", "User "+strconv.Itoa(i), "Password "+strconv.Itoa(i), "Email "+strconv.Itoa(i))
	}
}

func TestUpdateUser(t *testing.T) {
	addUsers(1)

	req, _ := http.NewRequest("GET", "/user/1", nil)
	response := executeRequest(req)
	originalUser := unmarshal(t, response.Body)

	payload := []byte(`{"name":"test user - updated name","password":"54321","email:"newemail@website.com"}`)

	req, _ = http.NewRequest("PUT", "/user/1", bytes.NewBuffer(payload))
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	m := unmarshal(t, response.Body)

	if m["id"] != originalUser["id"] {
		t.Errorf("Expected the id to remain the same (%v). Got %v", originalUser["id"], m["id"])
	}

	if m["name"] == originalUser["name"] {
		t.Errorf("Expected the name to change from '%v' to '%v'. Got '%v'", originalUser["name"], m["name"], m["name"])
	}

	if m["password"] == originalUser["password"] {
		t.Errorf("Expected the password to change from '%v' to '%v'. Got '%v'", originalUser["password"], m["password"], m["password"])
	}

	if m["email"] == originalUser["email"] {
		t.Errorf("Expected the password to change from '%v' to '%v'. Got '%v'", originalUser["email"], m["email"], m["email"])
	}
}

func TestDeleteUser(t *testing.T) {
	addUsers(1)

	req, _ := http.NewRequest("GET", "/user/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/user/1", nil)
	response = executeRequest(req)

	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/user/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func ensureTableExists() {
	if _, err := s.DB.Exec(tableCreationQuery); err != nil {
		//TODO: if no result for SHOW TABLES LIKE 'yourtable'; then create
		//for now we'll just run the query every time, if it doesn't exist it will make, if it does it will error but whatever
		//log.Fatal(err)
	}
}

func clearTable() {
	s.DB.Exec("DELETE FROM users")
	s.DB.Exec("ALTER SEQUENCE users_id_seq RESTART WITH 1")
}

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
    PRIMARY KEY (id)
)`
