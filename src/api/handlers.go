package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/IWannaCommunity/gate-jump/src/api/res"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest is the request expected on /login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// is the server alive
func (s *Server) getAlive(w http.ResponseWriter, r *http.Request) {
	res.New(http.StatusOK).JSON(w)
}

// get
func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid User ID").Error(w)
		return
	}

	u := User{ID: int64(id)}

	auth, response := s.getAuthLevel(r, &u)
	if response != nil {
		response.Error(w)
		return
	}

	if serr := u.getUser(s.DB, auth); serr.Err != nil {
		switch serr.Err {
		case sql.ErrNoRows:
			res.New(http.StatusNotFound).SetErrorMessage("User Not Found").Error(w)
		default:
			res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		}
		return
	}
	res.New(http.StatusOK).SetUser(u).JSON(w)
}

// get via name
func (s *Server) getUserByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	u := User{Name: &name}

	if serr := u.getUser(s.DB, PUBLIC); serr.Err != nil {
		switch serr.Err {
		case sql.ErrNoRows:
			res.New(http.StatusNotFound).SetErrorMessage("User Not Found").Error(w)
		default:
			res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		}
		return
	}

	res.New(http.StatusOK).SetUser(u).JSON(w)
}

// get multiple
func (s *Server) getUsers(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 50
	}
	if start < 0 {
		start = 0
	}

	auth, response := s.getAuthLevel(r, nil)
	if response != nil {
		response.Error(w)
		return
	}

	users, serr := getUsers(s.DB, start, count, auth)
	if serr != nil {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}

	res.New(http.StatusOK).SetUsers(users).JSON(w)
}

// register
func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var u User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid Request Payload").Error(w)
		return
	}
	defer r.Body.Close()
	//check if user with name already exists; if so, we will get an ErrNoRows which is what we want
	checkuser := u
	serr := checkuser.GetUserByName(s.DB)
	if serr.Err == nil {
		res.New(http.StatusBadRequest).SetErrorMessage("User Already Exists").Error(w)
		return
	} else if serr.Err != sql.ErrNoRows {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}

	//hash the password
	hashpwd, err := bcrypt.GenerateFromPassword([]byte(*u.Password), 12)
	if err != nil {
		res.New(http.StatusInternalServerError).SetErrorMessage("Failed Encrypting Password").Error(w)
		return
	}
	*u.Password = string(hashpwd)

	if serr := u.createUser(s.DB); serr != nil {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}

	res.New(http.StatusCreated).JSON(w)
}

// update
func (s *Server) updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid User ID").Error(w)
		return
	}

	var u User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid Request Payload").Error(w)
		return
	}
	defer r.Body.Close()
	u.ID = int64(id) // set expected id to url id value

	// get auth level of the request for the given id
	auth, response := s.getAuthLevel(r, &u)
	if response != nil {
		response.Error(w)
		return
	}
	// api requests made by permissions less than users can't edit any other user so reject them completely
	if auth < USER {
		res.New(http.StatusUnauthorized).SetErrorMessage("Requires User Permissions").Error(w)
		return
	}

	if u.Name != nil || u.Email != nil { // confirm unique email and username

	}

	//hash the password
	hashpwd, err := bcrypt.GenerateFromPassword([]byte(*u.Password), 12)
	if err != nil {
		res.New(http.StatusInternalServerError).SetErrorMessage("Failed Encrypting Password").Error(w)
		return
	}
	*u.Password = string(hashpwd)

	serr := u.updateUser(s.DB, auth)
	if serr.Err != nil {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}

	res.New(http.StatusOK).SetUser(u).JSON(w)
}

// delete
func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid User ID").Error(w)
		return
	}

	u := User{ID: int64(id)}

	auth, response := s.getAuthLevel(r, &u)
	if response != nil {
		response.Error(w)
		return
	}
	if auth < USER { // they arent the given user
		res.New(http.StatusUnauthorized).SetErrorMessage("Invalid Permissions").Error(w)
		return
	}
	if serr := u.deleteUser(s.DB); serr.Err != nil {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}
	res.New(http.StatusAccepted).JSON(w)
}

// login
func (s *Server) validateUser(w http.ResponseWriter, r *http.Request) {
	var lr LoginRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&lr); err != nil {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid Request Payload").Error(w)
		return
	}
	defer r.Body.Close()

	var u User
	u.Name = &lr.Username

	//get the user; if no user by that name, return 401, if other error, 500
	if serr := u.GetUserByName(s.DB, SERVER); serr.Err != nil {
		if serr.Err == sql.ErrNoRows {
			res.New(http.StatusUnauthorized).SetErrorMessage("Invalid Account").Error(w)
			return
		} else if serr.Err != nil {
			res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
			return
		}
	}
	//check the password
	if err := bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(lr.Password)); err != nil {
		res.New(http.StatusInternalServerError).SetErrorMessage("Failed Decrypting Password").Error(w)
		return
	}

	signedToken, err := u.CreateToken()
	if err != nil {
		res.New(http.StatusBadRequest).SetErrorMessage("Failed Creating Token").Error(w)
		return
	}

	u.LastToken = &signedToken
	u.LastLogin = &[]time.Time{time.Now()}[0] // how to get pointer from function call (its gross): goo.gl/9BXtsj
	u.LastIP = &r.RemoteAddr
	if serr := u.updateUser(s.DB, SERVER); serr.Err != nil {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}

	res.New(http.StatusOK).SetToken(signedToken).JSON(w)
}

func (s *Server) refreshUser(w http.ResponseWriter, r *http.Request) {
	// get user
	claims := r.Context().Value(CLAIMS).(Claims) // claims at this point are validated so refresh is allowed
	var u User
	u.ID = claims.ID
	if serr := u.getUser(s.DB, SERVER); serr.Err != nil {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}

	// make token
	token, err := u.CreateToken()
	if err != nil {
		res.New(http.StatusInternalServerError).SetErrorMessage("Error Creating Token").Error(w)
		return
	}

	// update login information
	u.LastToken = &token
	u.LastLogin = &[]time.Time{time.Now()}[0] // how to get pointer from function call (its gross): goo.gl/9BXtsj
	u.LastIP = &r.RemoteAddr

	// update information
	if serr := u.updateUser(s.DB, SERVER); serr.Err != nil {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}

	// return token
	res.New(http.StatusOK).SetToken(token).JSON(w)
}

// provide with request and said user and claims and confirm claims user exists and claims user's authentication level
func (s *Server) getAuthLevel(r *http.Request, u1 *User) (AuthLevel, *res.Response) {
	ctx := r.Context().Value(CLAIMS).(Context) // confirmed valid on jwt layer

	if ctx.claims.ID == 0 { // no claims exist
		return PUBLIC, nil
	}

	var u2 User
	u2.ID = ctx.claims.ID
	serr := u2.getUser(s.DB, SERVER)
	if serr.Err == sql.ErrNoRows { // claims user wasn't found
		return PUBLIC, res.New(http.StatusUnauthorized).SetErrorMessage("Token's User Doesn't Exist")
	} else if serr.Err != nil {
		return PUBLIC, res.New(http.StatusInternalServerError).SetInternalError(serr)
	}
	if u2.LastToken == nil || ctx.token != *u2.LastToken { // confirm the token is actually the last token used by the user
		return PUBLIC, res.New(http.StatusUnauthorized).SetErrorMessage("Token's User And Found User's Last Token Are Not The Same")
	}

	// we assume the username of the claimed user and the found user (u2) is the same because we searched by name
	if u2.Banned != nil && *u2.Banned { // fuck this guy in particular
		return PUBLIC, nil
	}

	if u1 == nil { // we aren't editing a user directly so no user was provided
		if u2.Admin != nil && *u2.Admin { // u2 is an admin
			return ADMIN, nil
		} else { // u2 is not an admin
			return PUBLIC, nil
		}
	} else {
		if u1.ID == u2.ID { // is u1 u2?
			if u2.Admin != nil && *u2.Admin { // is u2 an admin like they say they are?
				return ADMINUSER, nil
			} else { // u2 is not an admin but is u1
				return USER, nil
			}
		} else if u2.Admin != nil && *u2.Admin { // is u2 not u1 but is an admin?
			return ADMIN, nil
		}
		return PUBLIC, nil // u2 is neither u1 or an admin
	}
}
