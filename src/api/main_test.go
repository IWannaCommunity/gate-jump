package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	main "gate-jump/src/api"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

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
	s.DB.Exec("ALTER TABLE users AUTO_INCREMENT = 1")
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
    deleted BOOL NOT NULL DEFAULT FALSE,
    date_deleted DATETIME,
    PRIMARY KEY (id)
)`

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
	clearTable()
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

func TestCreateUser(t *testing.T) {
	clearTable()
	mainUser := []byte(`{"name":"test_user","password":"12345","email":"email@website.com"}`)
	duplicateName := []byte(`{"name":"test_user","password":"12345","email":"email@someotherwebsite.com"}`) // name == mainUser.Name
	duplicateEmail := []byte(`{"name":"some_other_user","password":"12345","email":"email@website.com"}`)   // email == mainUser.Email

	// test that you can create users
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(mainUser))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)
	m := unmarshal(t, response.Body)
	if m["success"] != true {
		t.Errorf("Expected 'true' got '%v'", m["success"])
	}
	// test that you can't create a user with a duplicate name
	req, _ = http.NewRequest("POST", "/register", bytes.NewBuffer(duplicateName))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusConflict, response.Code)
	m = unmarshal(t, response.Body)
	if m["success"] != false {
		t.Errorf("Expected 'false' got '%v'", m["success"])
	}
	if m["error"] != nil {
		if val := m["error"].(map[string]interface{})["message"].(string); val != "User Already Exists" {
			t.Errorf("Expected 'User Already Exists' got '%v'", val)
		}
	} else {
		t.Errorf("Expected something got '%v'", m["error"])
	}
	// test that you can't create a user with a duplicate email
	req, _ = http.NewRequest("POST", "/register", bytes.NewBuffer(duplicateEmail))
	response = executeRequest(req)
	checkResponseCode(t, http.StatusConflict, response.Code)
	m = unmarshal(t, response.Body)
	if m["success"] != false {
		t.Errorf("Expected 'false' got '%v'", m["success"])
	}
	if val := m["error"].(map[string]interface{})["message"].(string); val != "Email Already In Use" {
		t.Errorf("Expected 'Email Already In Use' got '%v'", val)
	}
}

func CreateUsers(count int) {
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

func TestLoginUser(t *testing.T) {
	clearTable()
	CreateUsers(1)
	wrongRe := []byte(`{"BADREQUEST"}`)
	wrongUP := []byte(`{"username":"wrongusername","password":"wrongpassword"}`)
	wrongUs := []byte(`{"username":"wrongusername","password":"password1"}`)
	wrongPa := []byte(`{"username":"user1","password":"wrongpassword"}`)
	correct := []byte(`{"username":"user1","password":"password1"}`)

	// test bad request
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(wrongRe))
	response := executeRequest(req)
	m := unmarshal(t, response.Body)
	checkResponseCode(t, http.StatusBadRequest, response.Code)
	if m["success"] != false {
		t.Errorf("Expected 'false' got '%v'", m["success"])
	}
	if m["error"] == nil {
		t.Errorf("Expected something got '%v'", m["error"])
	} else {
		if val := m["error"].(map[string]interface{})["message"].(string); val != "Invalid Request Payload" {
			t.Errorf("Expected 'Invalid Request Payload' got '%v'", val)
		}
	}

	// test wrong username and password
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(wrongUP))
	response = executeRequest(req)
	m = unmarshal(t, response.Body)
	checkResponseCode(t, http.StatusUnauthorized, response.Code)
	if m["success"] != false {
		t.Errorf("Expected 'false' got '%v'", m["success"])
	}
	if m["error"] == nil {
		t.Errorf("Expected something got '%v'", m["error"])
	} else {
		if val := m["error"].(map[string]interface{})["message"].(string); val != "User Doesn't Exist" {
			t.Errorf("Expected 'User Doesn't Exist' got '%v'", val)
		}
	}

	// test wrong username correct password
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(wrongUs))
	response = executeRequest(req)
	m = unmarshal(t, response.Body)
	checkResponseCode(t, http.StatusUnauthorized, response.Code)
	if m["success"] != false {
		t.Errorf("Expected 'false' got '%v'", m["success"])
	}
	if m["error"] == nil {
		t.Errorf("Expected something got '%v'", m["error"])
	} else {
		if val := m["error"].(map[string]interface{})["message"].(string); val != "User Doesn't Exist" {
			t.Errorf("Expected 'User Doesn't Exist' got '%v'", val)
		}
	}

	// test correct username wrong password
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(wrongPa))
	response = executeRequest(req)
	m = unmarshal(t, response.Body)
	checkResponseCode(t, http.StatusUnauthorized, response.Code)
	if m["success"] != false {
		t.Errorf("Expected 'false' got '%v'", m["success"])
	}
	if m["error"] == nil {
		t.Errorf("Expected something got '%v'", m["error"])
	} else {
		if val := m["error"].(map[string]interface{})["message"].(string); val != "Wrong Password" {
			t.Errorf("Expected 'Wrong Password' got '%v'", val)
		}
	}

	// test correct username correct password
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(correct))
	response = executeRequest(req)
	m = unmarshal(t, response.Body)
	checkResponseCode(t, http.StatusOK, response.Code)
	if m["success"] != true {
		t.Errorf("Expected 'true' got '%v'", m["success"])
	} else {
		if m["token"] == "" {
			t.Errorf("Expected something got '%v'", m["token"])
		}
	}
}

func loginUser(t *testing.T, username string, password string) string {
	payload_string := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)
	payload := []byte(payload_string)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(payload))
	response := executeRequest(req)
	result := unmarshal(t, response.Body)
	return result["token"].(string)
}
func TestGetUser(t *testing.T) {
	clearTable()

	// test getting non existent user
	req, _ := http.NewRequest("GET", "/user/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
	m := unmarshal(t, response.Body)
	if m["success"] != false {
		t.Errorf("Expected 'false' got '%v'", m["success"])
	}
	if val := m["error"].(map[string]interface{})["message"].(string); val != "User Not Found" {
		t.Errorf("Expected 'User Not Found' got '%v'", val)
	}

	CreateUsers(1)

	// test getting existing user
	req, _ = http.NewRequest("GET", "/user/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	m = unmarshal(t, response.Body)
	if val := m["success"]; val != true {
		t.Errorf("Expected 'true' got '%v'", val)
	}
	if val := m["user"]; val == nil {
		t.Errorf("Expected something got '%v'", val)
	} else {
		if user := m["user"].(map[string]interface{}); user == nil {
			t.Errorf("Expected something got '%v'", user)
		} else {
			if user["id"].(float64) != 1.0 {
				t.Errorf("Expected '1' got '%v'", user["id"])
			}
			if val := user["name"].(string); val != "user1" {
				t.Errorf("Expected 'user1' got '%v'", val)
			}
			if val := user["country"].(string); val != "us" {
				t.Errorf("Expected 'us' got '%v'", val)
			}
			if val := user["locale"].(string); val != "en" {
				t.Errorf("Expected 'en' got '%v'", val)
			}
		}
	}
}

func TestUpdateUserNoAuth(t *testing.T) {
	// check if auth is required
	payload := []byte(`{"name":"test user - updated name","password":"54321","email":"newemail@website.com","country":"jp","locale":"jp"}`)
	req, _ := http.NewRequest("PUT", "/user/1", bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusUnauthorized, response.Code)
}

func TestUpdateUser(t *testing.T) {
	payload := []byte(`{"name":"test user - updated name","password":"54321","email":"newemail@website.com","country":"jp","locale":"en"}`)
	req, _ := http.NewRequest("PUT", "/user/1", bytes.NewBuffer(payload))
	token := loginUser(t, "test user", "12345")
	req.Header.Set("Authorization", token)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	m := unmarshal(t, response.Body)
	if m["success"] != true {
		t.Errorf("Expected 'true' got '%v'", m["success"])
	} else {
		if val := m["user"].(map[string]interface{})["id"]; val != 1.0 {
			t.Errorf("Expected '1' got '%v'", val)
		}
		if val := m["user"].(map[string]interface{})["name"]; val != "test user - updated name" {
			t.Errorf("Expected 'test user - updated name' got '%v'", val)
		}
		if val := m["user"].(map[string]interface{})["password"]; val != nil {
			t.Errorf("Expected '<nil>' got '%v'", val)
		}
		if val := m["user"].(map[string]interface{})["email"]; val != "newemail@website.com" {
			t.Errorf("Expected 'newemail@website.com' got '%v'", val)
		}
		if val := m["user"].(map[string]interface{})["country"]; val != "jp" {
			t.Errorf("Expected 'jp' got '%v'", val)
		}
		if val := m["user"].(map[string]interface{})["locale"]; val != "en" {
			t.Errorf("Expected 'en' got '%v'", val)
		}
	}
}

func TestDeleteUserNoAuth(t *testing.T) {
	req, _ := http.NewRequest("GET", "/user/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/user/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusUnauthorized, response.Code)

	req, _ = http.NewRequest("GET", "/user/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestDeleteUser(t *testing.T) {
	req, _ := http.NewRequest("GET", "/user/1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("DELETE", "/user/1", nil)
	token := loginUser(t, "test user - updated name", "54321")
	req.Header.Set("Authorization", token)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusAccepted, response.Code)

	req, _ = http.NewRequest("GET", "/user/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestGetUsers(t *testing.T) {
	clearTable()
	count := 75
	CreateUsers(count)
	count = 50

	req, _ := http.NewRequest("GET", "/user", nil)
	response := executeRequest(req)
	m := unmarshal(t, response.Body)
	checkResponseCode(t, http.StatusOK, response.Code)
	if userList := m["userList"].(map[string]interface{}); m["success"] != true {
		t.Errorf("Expected 'true' got '%v'", m["success"])
	} else {
		var user map[string]interface{}
		users := userList["users"].([]interface{})
		if userList["startIndex"] != 0.0 {
			t.Errorf("Expected '0' got '%v'", userList["startingIndex"])
		}
		if userList["totalItems"] != float64(count) {
			t.Errorf("Expected '%v' got '%v'", float64(count), userList["totalItems"])
		}
		for i := 0; i < count; i++ {
			user = users[i].(map[string]interface{})
			if user["id"] != float64(i+1) {
				t.Errorf("Expected `%v` got `%v`", float64(i+1), user["id"])
			}
			if user["name"] != fmt.Sprintf("User %d", i) {
				t.Errorf("Expected `%v` got `%v`", fmt.Sprintf("User %d", i), user["name"])
			}
		}
	}
}
