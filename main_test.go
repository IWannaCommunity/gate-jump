package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	main "gate-jump"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
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

// confirm interacting with an empty user works as expected
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

// confirm creating user succeeds
func TestCreateUser(t *testing.T) {
	payload := []byte(`{"name":"test user","password":"12345","email":"email@website.com","country":"us","locale":"en"}`)
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(payload))
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)
	m := unmarshal(t, response.Body)
	if m["success"] != true {
		t.Errorf("Expected 'true' got '%v'", m["success"])
	}
}

func TestGetUser(t *testing.T) {
	// test getting non existent user
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
	// test getting existing user (from testcreateuser)
	req, _ = http.NewRequest("GET", "/user/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	result = unmarshal(t, response.Body)
	if result["success"] != true {
		t.Errorf("Expected 'true' got '%v'", result["success"])
	}
	if result["user"] == nil {
		t.Errorf("Expected something got '%v'", result["user"])
	} else {
		if user := result["user"].(map[string]interface{}); user == nil {
			t.Errorf("Expected something got '%v'", user)
		} else {
			if user["id"].(float64) != 1.0 {
				t.Errorf("Expected something got '%v'", user["id"])
			}
			if user["name"].(string) != "test user" {
				t.Errorf("Expected 'test user' got '%v'", user["name"].(string))
			}
			if user["country"] == "" {
				t.Errorf("Expected '<nil>' got '%v'", user["country"])
			}
			if user["locale"] == "" {
				t.Errorf("Expected '<nil>' got '%v'", user["locale"])
			}
		}
	}
}

func addUsers(count int) {
	if count < 1 {
		count = 1
	}

	for i := 0; i < count; i++ {
		s.DB.Exec("INSERT INTO users(name, password, email) VALUES(?, ?, ?)", "User "+strconv.Itoa(i), "Password "+strconv.Itoa(i), "Email "+strconv.Itoa(i))
	}
}

func TestLoginUser(t *testing.T) {
	payload := []byte(`{"username":"test user","password":"12345"}`)
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(payload))
	response := executeRequest(req)
	result := unmarshal(t, response.Body)
	if result["success"] != true {
		t.Errorf("Expected 'true' got '%v'", result["success"])
	} else {
		if result["token"] == "" {
			t.Errorf("Expected something got '%v'", result["token"])
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
	checkResponseCode(t, http.StatusOK, response.Code)

	req, _ = http.NewRequest("GET", "/user/1", nil)
	response = executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestGetUsers(t *testing.T) {
	clearTable()
	count := 75
	addUsers(count)
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
