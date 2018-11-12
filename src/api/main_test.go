package main_test

/*
func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func ensureTableExists() {
	if _, err := s.Exec(tableCreationQuery); err != nil {
		//TODO: if no result for SHOW TABLES LIKE 'yourtable'; then create
		//for now we'll just run the query every time, if it doesn't exist it will make, if it does it will error but whatever
		//log.Fatal(err)
	}
}

func clearTable() {
	db.Exec("DELETE FROM users")
	db.Exec("ALTER TABLE users AUTO_INCREMENT = 1")
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

var s *sql.DB
var r *mux.Router

// unmarshal will create a map from a json response, for quick and dirty testing
func unmarshal(t *testing.T, responseBody *bytes.Buffer) map[string]interface{} {
	var result map[string]interface{}
	if err := json.Unmarshal(responseBody.Bytes(), &result); err != nil {
		t.Error(err)
		return nil
	}
	return result
}

// ensures that the server is running before any tests begin as well as an empty table is provided for testing
func TestMain(m *testing.M) {
	//init server

	settings.FromFile("config/config.json")

	var err error
	s, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8mb4&parseTime=True&interpolateParams=true", settings.Database.Username, settings.Database.Password, settings.Database.Dsn))
	if err != nil {
		log.Fatal(err)
	}

	ensureTableExists()
	clearTable()

	code := m.Run()

	os.Exit(code)
}

// this should probably be baked into the testgetusers
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

func updateUser(id int, country string, locale string, admin bool, banned bool, deleted bool) {
	if deleted {
		s.Exec("UPDATE users SET country=?, locale=?, admin=?, banned=?, deleted=true, date_deleted=? WHERE id=?", country, locale, admin, banned, time.Now(), id)
	} else {
		s.Exec("UPDATE users SET country=?, locale=?, admin=?, banned=?, deleted=false WHERE id=?", country, locale, admin, banned, id)
	}
}

func TestLoginUser(t *testing.T) {
	clearTable()
	CreateUsers(2)
	wrongRe := []byte(`{"BADREQUEST"}`)
	wrongUP := []byte(`{"username":"wrongusername","password":"wrongpassword"}`)
	wrongUs := []byte(`{"username":"wrongusername","password":"password1"}`)
	wrongPa := []byte(`{"username":"user1","password":"wrongpassword"}`)
	correct := []byte(`{"username":"user1","password":"password1"}`)
	deleted := []byte(`{"username":"user2","password":"password2"}`)
	UpdateUser(2, "", "", false, false, true)

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

	// test correct username correct password deleted true
	req, _ = http.NewRequest("POST", "/login", bytes.NewBuffer(deleted))
	response = executeRequest(req)
	m = unmarshal(t, response.Body)
	checkResponseCode(t, http.StatusOK, response.Code)
	if m["success"] != true {
		t.Errorf("Expected 'true' got '%v'", m["success"])
	}
	if m["error"] != nil {
		t.Errorf("Expected 'nil' got '%v'", m["error"])
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
	CreateUsers(8)

	delete_token := loginUser(t, "user2", "password2") // expected public permissions (must get this before setting deleted=true or else account restored)
	banned_token := loginUser(t, "user3", "password3") // expected public permissions (must get this before setting banned=true or else can't login)

	UpdateUser(1, "us", "en", false, false, false) // just a user
	UpdateUser(2, "us", "en", false, false, true)  // deleted
	UpdateUser(3, "us", "en", false, true, false)  // banned
	UpdateUser(4, "us", "en", false, true, true)   // banned and deleted
	UpdateUser(5, "us", "en", true, false, false)  // admin
	UpdateUser(6, "us", "en", true, false, true)   // admin and deleted
	UpdateUser(7, "us", "en", true, true, false)   // admin and banned
	UpdateUser(8, "us", "en", true, true, true)    // admin and banned and deleted

	auser_token := loginUser(t, "user1", "password1")  // expected public permissions except for user1
	admini_token := loginUser(t, "user5", "password5") // expected admin permissions except for user5

	log.Println(banned_token) // to prevent it from erorr but maintaining login
	log.Println(admini_token) // to prevent it from error but maintaining login
	log.Println(delete_token) // to prevent it from error but maintaining login

	// test getting non existent user
	req, _ := http.NewRequest("GET", "/user/99", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
	m := unmarshal(t, response.Body)
	if val := m["success"]; val != false {
		t.Errorf("Expected 'false' got '%v'", val)
	} else if val := m["error"].(map[string]interface{})["message"].(string); val != "User Not Found" {
		t.Errorf("Expected 'User Not Found' got '%v'", val)
	}

	// auser check (user checking other records including their own)
	for i := 1; i <= 8; i++ {
		// check public creds
		istr := strconv.Itoa(i)
		req, _ = http.NewRequest("GET", "/user/"+istr, nil)
		req.Header.Set("Authorization", auser_token)
		response = executeRequest(req)
		m = unmarshal(t, response.Body)
		switch i {
		case 1: // user is checking themself so should have user perms for viewing data
			checkResponseCode(t, http.StatusOK, response.Code)
			if val := m["success"]; val != true {
				t.Errorf("Expected 'true' got '%v'", val)
			} else {
				user := m["user"].(map[string]interface{})
				if user == nil {
					t.Errorf("Expected something got '%v'", val)
				} else {
					if val := user["id"].(float64); val != float64(i) {
						t.Errorf("Expected '%d' got '%v'", i, val)
					}
					if val := user["name"].(string); val != "user"+istr {
						t.Errorf("Expected 'user%d' got '%v'", i, val)
					}
					if val := user["password"]; val != nil {
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
					if val := user["email"].(string); val != "email"+istr+"@website.com" {
						t.Errorf("Expected 'email%d@website.com' got '%v'", i, val)
					}
					if val := user["country"].(string); val != "us" {
						t.Errorf("Expected 'us' got '%v'", val)
					}
					if val := user["locale"].(string); val != "en" {
						t.Errorf("Expected 'en' got '%v'", val)
					}
					if val := user["date_created"].(string); val == "" {
						t.Errorf("Expected something got '%v'", val)
					}
					if val := user["verified"]; val == nil {
						t.Errorf("Expected something got '%v'", val)
					}
					if val := user["banned"]; val == nil {
						t.Errorf("Expected something got '%v'", val)
					}
					if val := user["admin"]; val == nil {
						t.Errorf("Expected something got '%v'", val)
					}
					if val := user["last_token"]; val != nil { // admin only
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
					if i == 1 || i == 3 || i == 5 { // these users have logged in so they have last_login set.
						if val := user["last_login"].(string); val == "" {
							t.Errorf("Expected something got '%v'", val)
						}
					} else { // these users have not logged in so they have last_login set to nil
						if val := user["last_login"]; val != nil {
							t.Errorf("Expected '<nil>' got '%v'", val)
						}
					}
					if val := user["last_ip"]; val != nil { // admin only
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
					if val := user["deleted"]; val != nil { // admin only
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
					if val := user["date_deleted"]; val != nil { // admin only
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
				}
			}
		case 2: // user is checking a deleted user so should not be able to see anything
			fallthrough
		case 4:
			fallthrough
		case 6:
			fallthrough
		case 8:
			checkResponseCode(t, http.StatusNotFound, response.Code)
			if val := m["success"]; val != false {
				t.Errorf("Expected 'false' got '%v'", val)
			} else if val := m["error"].(map[string]interface{})["message"].(string); val != "User Not Found" {
				t.Errorf("Expected 'User Not Found' got '%v'", val)
			}
		case 3: // user is checking other users. this should be public
			fallthrough
		case 5:
			fallthrough
		case 7:
			checkResponseCode(t, http.StatusOK, response.Code)
			log.Println(m)
			if val := m["success"]; val != true {
				t.Errorf("Expected 'true' got '%v'", val)
			} else {
				user := m["user"].(map[string]interface{})
				if user == nil {
					t.Errorf("Expected something got '%v'", val)
				} else {
					if val := user["id"].(float64); val != float64(i) {
						t.Errorf("Expected '%d' got '%v'", i, val)
					}
					if val := user["name"].(string); val != "user"+istr {
						t.Errorf("Expected 'user%d' got '%v'", i, val)
					}
					if val := user["password"]; val != nil {
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
					if val := user["email"]; val != nil {
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
					if val := user["country"].(string); val != "us" {
						t.Errorf("Expected 'us' got '%v'", val)
					}
					if val := user["locale"].(string); val != "en" {
						t.Errorf("Expected 'en' got '%v'", val)
					}
					if val := user["date_created"].(string); val == "" {
						t.Errorf("Expected something got '%v'", val)
					}
					if val := user["verified"]; val == nil {
						t.Errorf("Expected something got '%v'", val)
					}
					if val := user["banned"]; val == nil {
						t.Errorf("Expected something got '%v'", val)
					}
					if val := user["admin"]; val == nil {
						t.Errorf("Expected something got '%v'", val)
					}
					if val := user["last_token"]; val != nil { // admin only
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
					if i == 1 || i == 3 || i == 5 { // these users have logged in so they have last_login set.
						if val := user["last_login"].(string); val == "" {
							t.Errorf("Expected something got '%v'", val)
						}
					} else { // these users have not logged in so they have last_login set to nil
						if val := user["last_login"]; val != nil {
							t.Errorf("Expected '<nil>' got '%v'", val)
						}
					}
					if val := user["last_ip"]; val != nil { // admin only
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
					if val := user["deleted"]; val != nil { // admin only
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
					if val := user["date_deleted"]; val != nil { // admin only
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
				}
			}
		}
	}

	// banned_token
	for i := 1; i <= 8; i++ {
		// check public creds
		istr := strconv.Itoa(i)
		req, _ = http.NewRequest("GET", "/user/"+istr, nil)
		req.Header.Set("Authorization", auser_token)
		response = executeRequest(req)
		m = unmarshal(t, response.Body)
		switch i {
		case 1: // user is checking themself so should have user perms for viewing data
			checkResponseCode(t, http.StatusOK, response.Code)
			if val := m["success"]; val != true {
				t.Errorf("Expected 'true' got '%v'", val)
			} else {
				user := m["user"].(map[string]interface{})
				if user == nil {
					t.Errorf("Expected something got '%v'", val)
				} else {
					if val := user["id"].(float64); val != float64(i) {
						t.Errorf("Expected '%d' got '%v'", i, val)
					}
					if val := user["name"].(string); val != "user"+istr {
						t.Errorf("Expected 'user%d' got '%v'", i, val)
					}
					if val := user["password"]; val != nil {
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
					if val := user["email"].(string); val != "email"+istr+"@website.com" {
						t.Errorf("Expected 'email%d@website.com' got '%v'", i, val)
					}
					if val := user["country"].(string); val != "us" {
						t.Errorf("Expected 'us' got '%v'", val)
					}
					if val := user["locale"].(string); val != "en" {
						t.Errorf("Expected 'en' got '%v'", val)
					}
					if val := user["date_created"].(string); val == "" {
						t.Errorf("Expected something got '%v'", val)
					}
					if val := user["verified"]; val == nil {
						t.Errorf("Expected something got '%v'", val)
					}
					if val := user["banned"]; val == nil {
						t.Errorf("Expected something got '%v'", val)
					}
					if val := user["admin"]; val == nil {
						t.Errorf("Expected something got '%v'", val)
					}
					if val := user["last_token"]; val != nil { // admin only
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
					if i == 1 || i == 3 || i == 5 { // these users have logged in so they have last_login set.
						if val := user["last_login"].(string); val == "" {
							t.Errorf("Expected something got '%v'", val)
						}
					} else { // these users have not logged in so they have last_login set to nil
						if val := user["last_login"]; val != nil {
							t.Errorf("Expected '<nil>' got '%v'", val)
						}
					}
					if val := user["last_ip"]; val != nil { // admin only
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
					if val := user["deleted"]; val != nil { // admin only
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
					if val := user["date_deleted"]; val != nil { // admin only
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
				}
			}
		case 2: // user is checking a deleted user so should not be able to see anything
			fallthrough
		case 4:
			fallthrough
		case 6:
			fallthrough
		case 8:
			checkResponseCode(t, http.StatusNotFound, response.Code)
			if val := m["success"]; val != false {
				t.Errorf("Expected 'false' got '%v'", val)
			} else if val := m["error"].(map[string]interface{})["message"].(string); val != "User Not Found" {
				t.Errorf("Expected 'User Not Found' got '%v'", val)
			}
		case 3: // user is checking other users. this should be public
			fallthrough
		case 5:
			fallthrough
		case 7:
			checkResponseCode(t, http.StatusOK, response.Code)
			log.Println(m)
			if val := m["success"]; val != true {
				t.Errorf("Expected 'true' got '%v'", val)
			} else {
				user := m["user"].(map[string]interface{})
				if user == nil {
					t.Errorf("Expected something got '%v'", val)
				} else {
					if val := user["id"].(float64); val != float64(i) {
						t.Errorf("Expected '%d' got '%v'", i, val)
					}
					if val := user["name"].(string); val != "user"+istr {
						t.Errorf("Expected 'user%d' got '%v'", i, val)
					}
					if val := user["password"]; val != nil {
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
					if val := user["email"]; val != nil {
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
					if val := user["country"].(string); val != "us" {
						t.Errorf("Expected 'us' got '%v'", val)
					}
					if val := user["locale"].(string); val != "en" {
						t.Errorf("Expected 'en' got '%v'", val)
					}
					if val := user["date_created"].(string); val == "" {
						t.Errorf("Expected something got '%v'", val)
					}
					if val := user["verified"]; val == nil {
						t.Errorf("Expected something got '%v'", val)
					}
					if val := user["banned"]; val == nil {
						t.Errorf("Expected something got '%v'", val)
					}
					if val := user["admin"]; val == nil {
						t.Errorf("Expected something got '%v'", val)
					}
					if val := user["last_token"]; val != nil { // admin only
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
					if i == 1 || i == 3 || i == 5 { // these users have logged in so they have last_login set.
						if val := user["last_login"].(string); val == "" {
							t.Errorf("Expected something got '%v'", val)
						}
					} else { // these users have not logged in so they have last_login set to nil
						if val := user["last_login"]; val != nil {
							t.Errorf("Expected '<nil>' got '%v'", val)
						}
					}
					if val := user["last_ip"]; val != nil { // admin only
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
					if val := user["deleted"]; val != nil { // admin only
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
					if val := user["date_deleted"]; val != nil { // admin only
						t.Errorf("Expected '<nil>' got '%v'", val)
					}
				}
			}
		}
	}

}

func TestOldToken(t *testing.T) {
	// confirm that old tokens are rejected by default
	clearTable()
	CreateUsers(1)

	old_token := loginUser(t, "user1", "password1") // get a token
	_ = loginUser(t, "user1", "password1")          // get server to update the last_token field to something new

	req, _ := http.NewRequest("GET", "/user/1", nil)
	req.Header.Set("Authorization", old_token)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusUnauthorized, response.Code)
	m := unmarshal(t, response.Body)
	log.Println(m)
	if val := m["success"]; val != false {
		t.Errorf("Expected 'false' got '%v'", val)
	}
}

func TestUpdateUser(t *testing.T) {
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
*/
