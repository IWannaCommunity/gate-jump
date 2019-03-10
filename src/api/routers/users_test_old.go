package routers

/*
import (
	tst "github.com/IWannaCommunity/gate-jump/src/api/testing"
)

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

var te tst.TestingEnv

func TestMain(m *testing.M) {
	var err error
	// initialize the database for testing
	//settings.FromFile("config/config.json")

	if err = settings.FromFile("/home/dev/go/src/github.com/IWannaCommunity/gate-jump/src/api/config/config.json"); err != nil {
		log.Fatal("Failed loading configuration: ", err)
	}

	database.Connect(settings.Database.Username, settings.Database.Password, settings.Database.Dsn)

	// initialize the tests database connection
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8mb4&parseTime=True&interpolateParams=true", settings.Database.Username, settings.Database.Password, settings.Database.Dsn))
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
	var r TestPayload
	r = request("GET", "/", nil, nil, nil)
	if assert.NoError(t, r.Error, Got(r, "api error"))) {

		assert.Equal(t, http.StatusOK, r.Code, Expected(r, "http.StatusOK")))
		assert.True(t, r.Response.Success, Expected(r, "successful request")))

		assert.Nil(t, r.Response.Error, Got(r, "error")))
		assert.Nil(t, r.Response.Token, Got(r, "token")))
		assert.Nil(t, r.Response.User, Got(r, "user")))
		assert.Nil(t, r.Response.UserList, Got(r, "userlist")))

	}
}

func TestCreateUser(t *testing.T) {
	clearTable()
	var r TestPayload
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
		r = request(method, route, badRequest, nil, nil)
		if assert.NoError(t, r.Error, Got(r, "api request error")) {

			assert.Equal(t, http.StatusBadRequest, r.Code, Expected(r, "http.StatusBadRequest"))
			assert.False(t, r.Response.Success, Expected(r, "failed request"))
			if assert.NotNil(t, r.Response.Error, Got(r, "no request error")) {
				switch i {
				case 5:
					fallthrough
				case 6: // invalid username
					assert.Equal(t, "Invalid Username Format", *r.Response.Error, Expected(r, "Invalid Username Format"))
				case 7:
					assert.Equal(t, "Invalid Password Format", *r.Response.Error, Expected(r, "Invalid Password Format"))
				case 8:
					assert.Equal(t, "Invalid Email Format", *r.Response.Error, Expected(r, "Invalid Email Format"))
				default:
					assert.Equal(t, "Invalid Request Payload", *r.Response.Error, Expected(r, "Invalid Request Payload"))
				}
			}

			assert.Nil(t, r.Response.Token, Got(r, "token"))
			assert.Nil(t, r.Response.User, Got(r, "user"))
			assert.Nil(t, r.Response.UserList, Got(r, ""))
		}
	}

	// test creating a user
	r = request(method, route, mainUser, nil, nil)
	if assert.NoError(t, r.Error) {

		assert.Equal(t, http.StatusCreated, r.Code, Expected(r,"http.StatusCreated"))
		assert.True(t, r.Response.Success, Expected(r,"succsessful api request"))

		assert.Nil(t, r.Response.Error, Got("error"))
		assert.Nil(t, r.Response.Token, Got("token"))
		assert.Nil(t, r.Response.User, Got("user"))
		assert.Nil(t, r.Response.UserList, Got("userlist"))
	}
	// test duplicate username
	r = request(method, route, duplicateName, nil, nil)
	if assert.NoError(t, r.Error) {

		assert.Equal(t, http.StatusConflict, r.Code, Expected(r,"http.StatusConflict")
		assert.False(t, r.Response.Success,Expected(r,"succsessful api request"))
		if assert.NotNil(t, r.Response.Error) {
			assert.Equal(t, "Username Already Exists", *r.Response.Error)
		}

		assert.Nil(t, r.Response.Token)
		assert.Nil(t, r.Response.User)
		assert.Nil(t, r.Response.UserList)
	}

	// test duplicate email
	r = request(method, route, duplicateEmail, nil, nil)
	if assert.NoError(t, r.Error) {

		assert.Equal(t, http.StatusConflict, r.Code, "expected statusconflict")
		assert.False(t, r.Response.Success)
		if assert.NotNil(t, r.Response.Error) {
			assert.Equal(t, "Email Already In Use", *r.Response.Error)
		}

		assert.Nil(t, r.Response.Token)
		assert.Nil(t, r.Response.User)
		assert.Nil(t, r.Response.UserList)
	}

}

func TestLoginUser(t *testing.T) {
	clearTable()
	create(3)
	update(2, "us", "en", false, false, true) // deleted
	update(3, "us", "en", false, true, false) // banned
	var code int
	var r TestPayload
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
		r = request(method, route, badRequest, nil, nil)
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
			assert.Falsef(t, r.Response.Success, "badRequest: %d", i)
			if assert.NotNilf(t, r.Response.Error, "badRequest: %d", i) {
				switch i {
				case 0:
					fallthrough
				case 1:
					fallthrough
				case 2:
					assert.Equalf(t, "Invalid Request Payload", r.Response.Error, "badRequest: %d", i)
				case 3: // extra info, we dont care
					fallthrough
				case 4: // wrong username wrong password
					fallthrough
				case 5: // wrong
					assert.Equalf(t, "User Doesn't Exist", r.Response.Error, "badRequest: %d", i)
				case 6:
					assert.Equalf(t, "Wrong Password", r.Response.Error, "badRequest: %d", i)
				default:
					assert.Truef(t, false, "badRequest: %d", i)
					return // shouldn't occur
				}
			}

			assert.Nilf(t, r.Response.Token, "badRequest: %d", i)
			assert.Nilf(t, r.Response.User, "badRequest: %d", i)
			assert.Nilf(t, r.Response.UserList, "badRequest: %d", i)
		}
	}

	// correct username correct password; not banned; not deleted
	r = request(method, route, correct, nil, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, code, "expected OK")
		if assert.True(t, r.Response.Success) {
			assert.NotNil(t, r.Response.Token)
		}

		assert.Nil(t, r.Response.Error)
		assert.Nil(t, r.Response.User)
		assert.Nil(t, r.Response.UserList)
	}

	// correct username correct password; not banned; deleted
	r = request(method, route, deleted, nil, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, code, "expected OK")
		if assert.True(t, r.Response.Success) {
			assert.NotNil(t, r.Response.Token)
		}

		assert.Nil(t, r.Response.Error)
		assert.Nil(t, r.Response.User)
		assert.Nil(t, r.Response.UserList)
	}

	// correct username correct password; banned; not deleted
	r = request(method, route, banned, nil, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusUnauthorized, code, "expected unauthorized")
		if assert.False(t, r.Response.Success) {
			if assert.NotNil(t, r.Response.Error) {
				assert.Equal(t, "Account Banned", r.Response.Error)
			}
		}

		assert.Nil(t, r.Response.Token)
		assert.Nil(t, r.Response.User)
		assert.Nil(t, r.Response.UserList)
	}
}

func TestRefresh(t *testing.T) {

	clearTable()
	create(1)
	var code int
	var r TestPayload
	var err error
	method := "POST"
	route := "/refresh"

	var badRequests []string

	badRequests = append(badRequests,
		"badtoken")

	goodToken := login("user1", "password1")

	for i, token := range badRequests {
		r = request(method, route, nil, token, nil)
		if assert.NoErrorf(t, err, "badRequest %d", i) {
			assert.Equalf(t, http.StatusUnauthorized, code, "expected unauthorized", "badRequest %d", i)
			assert.Falsef(t, r.Response.Success, "badRequest %d", i)
			if assert.NotNilf(t, r.Response.Error, "badRequest %d", i) {
				assert.Equalf(t, "Invalid Token Provided", r.Response.Error, "badRequest %d", i)
			}

			assert.Nil(t, r.Response.Token, "badRequest %d", i)
			assert.Nil(t, r.Response.User, "badRequest %d", i)
			assert.Nil(t, r.Response.UserList, "badRequest %d", i)
		}
	}

	r = request(method, route, nil, goodToken, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, code, "expected OK")
		assert.True(t, r.Response.Success)
		assert.NotNil(t, r.Response.Token)

		assert.Nil(t, r.Response.Error)
		assert.Nil(t, r.Response.User)
		assert.Nil(t, r.Response.UserList)
	}
}

func TestGetUser(t *testing.T) {

	clearTable()
	create(10)
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
	checkNONEXISTINGUser(t, request(method, route+"99", nil, nil, nil), "public", route+"99")

	// search self (user)
	checkUSERCredentials(t, request(method, route+"9", nil, userToken, nil), "user", route+"9")

	// search self (admin)
	checkADMINCredentials(t, request(method, route+"10", nil, adminToken, nil), "admin", route+"10")

	for i := 0; i < 3; i++ {
		var currentToken interface{}
		var currentRoute string
		var currentType Level
		switch i {
		case 0:
			currentToken = nil
			currentType = PUBLIC
		case 1:
			currentToken = userToken
			currentType = USER
		case 2:
			currentToken = adminToken
			currentType = ADMIN
		}
		for id := 1; id <= 8; id++ {
			currentRoute = route + strconv.Itoa(id)
			switch id {
			case 1: // regular user
				checkUser(t, request(method, currentRoute, nil, currentToken, nil), currentType, currentRoute)
			case 3: // banned
				checkBANNEDUser(t, request(method, currentRoute, nil, currentToken, nil), currentType, currentRoute)
			case 5: // admin
				checkUser(t, request(method, currentRoute, nil, currentToken, nil), currentType, currentRoute)
			case 7: // admin and banned
				checkBANNEDUser(t, request(method, currentRoute, nil, currentToken, nil), currentType, currentRoute)
			case 2: // deleted
				checkDELETEDUser(t, request(method, currentRoute, nil, currentToken, nil), currentType, currentRoute)
			case 4: // banned and deleted
				checkDELETEDUser(t, request(method, currentRoute, nil, currentToken, nil), currentType, currentRoute)
			case 6: // admin and deleted
				checkDELETEDUser(t, request(method, currentRoute, nil, currentToken, nil), currentType, currentRoute)
			case 8: // admin and banned and deleted
				checkDELETEDUser(t, request(method, currentRoute, nil, currentToken, nil), currentType, currentRoute)
			}
		}
	}

}

func TestGetUsers(t *testing.T) {
	clearTable()

	var code int
	var r TestPayload
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
	r = request(method, route, nil, nil, defaultFormValue)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, code)
		assert.True(t, r.Response.Success)
		if assert.NotNil(t, r.Response.UserList) {
			assert.Equal(t, 0, r.Response.UserList.StartIndex)
			assert.Equal(t, 0, r.Response.UserList.TotalItems)
			assert.Nil(t, r.Response.UserList.Users)
		}

		assert.Nil(t, r.Response.Token)
		assert.Nil(t, r.Response.User)
		assert.Nil(t, r.Response.Error)
	}

	// test bad request
	r = request(method, route, nil, nil, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, code)
		assert.True(t, r.Response.Success)
		if assert.NotNil(t, r.Response.UserList) {
			assert.Equal(t, 0, r.Response.UserList.StartIndex)
			assert.Equal(t, 0, r.Response.UserList.TotalItems)
			assert.Nil(t, r.Response.UserList.Users)
		}

		assert.Nil(t, r.Response.Token)
		assert.Nil(t, r.Response.User)
		assert.Nil(t, r.Response.Error)
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

		r = request(method, route, nil, nil, form)
		assert.Equalf(t, http.StatusOK, code, "{start: %d; count: %d}", startGiven, countGiven)
		assert.Truef(t, r.Response.Success, "{start: %d; count: %d}", startGiven, countGiven)
		if assert.NotNilf(t, r.Response.UserList, "{start: %d; count: %d}", startGiven, countGiven) {
			assert.Equalf(t, start, r.Response.UserList.StartIndex, "{start: %d; count: %d}", startGiven, countGiven)
			assert.Equalf(t, count, r.Response.UserList.TotalItems, "{start: %d; count: %d}", startGiven, countGiven)
			if count > 0 {
				assert.NotNilf(t, r.Response.UserList.Users, "{start: %d; count: %d}", startGiven, countGiven)
			} else {
				assert.Nilf(t, r.Response.UserList.Users, "{start: %d; count: %d}", startGiven, countGiven)
			}

			assert.Nilf(t, r.Response.Token, "{start: %d; count: %d}", startGiven, countGiven)
			assert.Nilf(t, r.Response.User, "{start: %d; count: %d}", startGiven, countGiven)
			assert.Nilf(t, r.Response.Error, "{start: %d; count: %d}", startGiven, countGiven)
		}
	}
}

func TestGetUserByName(t *testing.T) {
	clearTable()
	create(10)
	var code int
	var r TestPayload
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

	update(9, "us", "en", true, false, false) // user perms for login
	update(10, "us", "en", true, true, true)  // admin user for login
	userToken := login("user9", "password9")
	adminToken := login("user10", "password10")

	// search for nonexisting user
	r = request(method, route+"ParagusRants", nil, nil, nil)
	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusNotFound, code)
		assert.False(t, r.Response.Success)
		if assert.NotNil(t, r.Response.Error) {
			assert.Equal(t, "User Not Found", r.Response.Error)
		}

		assert.Nil(t, r.Response.Token)
		assert.Nil(t, r.Response.User)
		assert.Nil(t, r.Response.UserList)
	}

	// search as non-token user and admin (should all be public)
	for i := 0; i < 3; i++ {
		var token interface{}
		var requestingUser string
		switch i {
		case 0:
			token = nil
			requestingUser = "non-user"
		case 1:
			token = userToken
			requestingUser = "user"
		case 2:
			token = adminToken
			requestingUser = "admin"
		}
		for id := 1; id <= 10; id++ {
			name := "user" + strconv.Itoa(id)
			r = request(method, route+name, nil, token, nil)
			if assert.NoError(t, err, "request %s; requester %s", name, requestingUser) {
				switch id {
				case 1: // regular user
					fallthrough
				case 3: // banned
					fallthrough
				case 5: // admin
					fallthrough
				case 7: // admin and banned
					assert.Equalf(t, http.StatusOK, code, "request %s; requester %s", name, requestingUser)
					assert.True(t, r.Response.Success, "request %s; requester %s", name, requestingUser)
					if assert.NotNilf(t, r.Response.User, "request %s; requester %s", name, requestingUser) {
						assert.NotNilf(t, r.Response.User.ID, "request %s; requester %s", name, requestingUser)
						assert.NotNilf(t, r.Response.User.Name, "request %s; requester %s", name, requestingUser)
						assert.Nilf(t, r.Response.User.Password, "request %s; requester %s", name, requestingUser)
						assert.Nilf(t, r.Response.User.Email, "request %s; requester %s", name, requestingUser)
						assert.NotNilf(t, r.Response.User.Country, "request %s; requester %s", name, requestingUser)
						assert.NotNilf(t, r.Response.User.Locale, "request %s; requester %s", name, requestingUser)
						assert.NotNilf(t, r.Response.User.DateCreated, "request %s; requester %s", name, requestingUser)
						assert.NotNilf(t, r.Response.User.Verified, "request %s; requester %s", name, requestingUser)
						assert.NotNilf(t, r.Response.User.Banned, "request %s; requester %s", name, requestingUser)
						assert.NotNilf(t, r.Response.User.Admin, "request %s; requester %s", name, requestingUser)
						assert.Nilf(t, r.Response.User.LastToken, "request %s; requester %s", name, requestingUser)
						assert.NotNilf(t, r.Response.User.LastLogin, "request %s; requester %s", name, requestingUser)
						assert.Nilf(t, r.Response.User.LastIP, "request %s; requester %s", name, requestingUser)
						assert.Nilf(t, r.Response.User.Deleted, "request %s; requester %s", name, requestingUser)
						assert.Nilf(t, r.Response.User.DateDeleted, "request %s; requester %s", name, requestingUser)
					}
					assert.Nilf(t, r.Response.Error, "request %s; requester %s", name, requestingUser)
				case 2: // deleted
					fallthrough
				case 4: // banned and deleted
					fallthrough
				case 6: // admin and deleted
					fallthrough
				case 8: // admin and banned and deleted
					assert.Equalf(t, http.StatusNotFound, code, "request %s; requester %s", name, requestingUser)
					assert.Falsef(t, r.Response.Success, "request %s; requester %s", name, requestingUser)
					if assert.NotNilf(t, r.Response.Error, "request %s; requester %s", name, requestingUser) {
						assert.Equal(t, "User Not Found", r.Response.Error, "request %s; requester %s", name, requestingUser)
					}
					assert.Nilf(t, r.Response.User, "request %s; requester %s", name, requestingUser)
				}
				assert.Nilf(t, r.Response.Token, "request %s; requester %s", name, requestingUser)
				assert.Nilf(t, r.Response.UserList, "request %s; requester %s", name, requestingUser)
			}

		}
	}

}

func TestUpdateUser(t *testing.T) {
	t.Error("Not Implemented")
}

func TestBanUser(t *testing.T) {
	t.Error("Not implemented")
}

func TestDeleteUser(t *testing.T) {
	clearTable()
	create(2)
	var r TestPayload
	route := "/user/" // add id to the end of this

	update(1, "us", "en", false, false, false) // the user to test
	update(2, "us", "en", true, false, false)  // the admin user that can verify user is deleted

	userToken := login("user1", "password1")  // user token
	adminToken := login("user2", "password2") // deleted token

	// check user exists in database before attempting to delete
	checkUser(t, request("GET", route+"1", nil, nil, nil), "public", route+"1")

	// delete user
	r = request("DELETE", route+"1", nil, adminToken, nil)
	if assert.NoError(t, r.Error, "Failed making delete user api request.") {
		assert.Equal(t, http.StatusAccepted, r.Code, "Expected http.StatusAccepted")
		assert.True(t, r.Response.Success, "Expected successful API request")
		assert.Nil(t, r.Response.Token, "GOT TOKEN?!")
		assert.Nil(t, r.Response.User, "GOT USER?!")
		assert.Nil(t, r.Response.UserList, "GOT USERLIST?!")
		assert.Nil(t, r.Response.Error, "GOT ERROR?!")
	}

	// check deleted user as public user
	checkDELETEDUser(t, request("GET", route+"1", nil, nil, nil), "public", route+"1")

	// check deleted user as admin user
	checkDELETEDUser(t, request("GET", route+"1", nil, adminToken, nil), "admin", route+"1")

	// check deleted user as the deleted user's token (should reject)
	checkNONEXISTINGUser(t, request("GET", route+"1", nil, userToken, nil), "user", route+"1")

	// login to deleted user to undelete it
	userToken = login("user1", "password1")

	/// check originally deleted user as public user
	checkUser(t, request("GET", route+"1", nil, nil, nil), "public", route+"1")

	// check originally deleted user as self
	checkUser(t, request("GET", route+"1", nil, userToken, nil), "user", route+"1")

}

func TestCheckUser(t *testing.T) {
	clearTable()
	create(4)

	update(1, "us", "en", false, false, false)
	update(2, "us", "en", true, false, false) // admin
	update(3, "us", "en", false, true, false) // banned
	update(4, "us", "en", false, false, true) // deleted

	adminToken := login("user2", "password2")

	// public
	checkUser(t, request("GET", "/user/1", nil, nil, nil), "public", "/user/1")                       // test public credentials
	checkUser(t, request("GET", "/user/1", nil, login("user1", "password1"), nil), "user", "/user/1") // test user credentials
	checkUser(t, request("GET", "/user/1", nil, adminToken, nil), "admin", "/user/1")                 // test admin credentials

	// banned
	//checkUser(t, request("GET", "/user/3", nil, nil, nil), "public", "/user/3")       // test public credentials
	//checkUser(t, request("GET", "/user/3", nil, adminToken, nil), "admin", "/user/3") // test admin credentials

	// deleted
	//checkUser(t, request("GET", "/user/4", nil, nil, nil), "public", "/user/4")       // test public credentials
	//checkUser(t, request("GET", "/user/4", nil, adminToken, nil), "admin", "/user/4") // test admin credentials

}

// HELPER FUNCTIONS FOR CHECKING RETURNED USER DATA ==============================================================
*/
