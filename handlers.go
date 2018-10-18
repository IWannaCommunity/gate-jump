package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) getAlive(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, NewResponse(http.StatusOK, "", map[string]string{"alive": "true"}))
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, NewResponse(http.StatusBadRequest, "Invalid user ID", nil))
		return
	}

	u := User{UserID: int64(id)}
	if err := u.getUser(s.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, NewResponse(http.StatusNotFound, "User not found", nil))
		default:
			response := NewResponse(http.StatusInternalServerError, "", nil)
			response.Err = err
			respondWithError(w, response)
		}
		return
	}
	respondWithJSON(w, NewResponse(http.StatusOK, "", u))
}

func (s *Server) getUsers(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 50
	}
	if start < 0 {
		start = 0
	}

	users, err := getUsers(s.DB, start, count)
	if err != nil {
		response := NewResponse(http.StatusInternalServerError, "", nil)
		response.Err = err
		respondWithError(w, response)
		return
	}

	respondWithJSON(w, NewResponse(http.StatusOK, "", users))
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var u User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, NewResponse(http.StatusBadRequest, "Invalid request payload", nil))
		return
	}
	defer r.Body.Close()

	//check if user with name already exists; if so, we will get an ErrNoRows which is what we want
	checkuser := u
	err := checkuser.getUserByName(s.DB)
	if err == nil {
		respondWithError(w, NewResponse(http.StatusBadRequest, "User already exists", nil))
		return
	} else if err != sql.ErrNoRows {
		response := NewResponse(http.StatusInternalServerError, "", nil)
		response.Err = err
		respondWithError(w, response)
		return
	}

	//hash the password
	hashpwd, err := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
	if err != nil {
		response := NewResponse(http.StatusInternalServerError, "Failed encrypting password", nil)
		response.Err = err
		respondWithError(w, response)
		return
	}
	u.Password = string(hashpwd)

	if err := u.createUser(s.DB); err != nil {
		response := NewResponse(http.StatusInternalServerError, "Failed creating user", nil)
		response.Err = err
		respondWithError(w, response)
		return
	}
	respondWithJSON(w, NewResponse(http.StatusCreated, "", u))
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, NewResponse(http.StatusBadRequest, "Invalid user ID", nil))
		return
	}

	var u User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, NewResponse(http.StatusBadRequest, "Invalid request payload", ""))
		return
	}
	defer r.Body.Close()
	u.UserID = int64(id)

	if err := u.updateUser(s.DB); err != nil {
		response := NewResponse(http.StatusInternalServerError, "", nil)
		response.Err = err
		respondWithError(w, response)
		return
	}
	respondWithJSON(w, NewResponse(http.StatusOK, "", u))
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, NewResponse(http.StatusBadRequest, "Invalid user ID", nil))
		return
	}

	u := User{UserID: int64(id)}
	if err := u.deleteUser(s.DB); err != nil {
		response := NewResponse(http.StatusInternalServerError, "", nil)
		response.Err = err
		respondWithError(w, response)
		return
	}

	respondWithJSON(w, NewResponse(http.StatusOK, "", map[string]string{"result": "success"}))
}

// LoginRequest is the request expected on /login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) validateUser(w http.ResponseWriter, r *http.Request) {
	var lr LoginRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&lr); err != nil {
		respondWithError(w, NewResponse(http.StatusBadRequest, "Invalid request payload", nil))
		return
	}
	defer r.Body.Close()

	var u User
	u.Username = lr.Username

	//get the user; if no user by that name, return 401, if other error, 500
	if err := u.getUserByName(s.DB); err != nil {
		if err == sql.ErrNoRows {

			respondWithError(w, NewResponse(http.StatusUnauthorized, "Invalid Account", nil))
			return
		} else {
			response := NewResponse(http.StatusInternalServerError, "", nil)
			response.Err = err
			respondWithError(w, response)
			return
		}
	}

	//check the password
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(lr.Password)); err != nil {
		respondWithError(w, NewResponse(http.StatusUnauthorized, "Invalid Account", nil))
		return
	}

	//create and sign the token
	claims := Claims{
		u.Username,
		u.Admin,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(), //expire in one hour
			Issuer:    Config.Host + ":" + Config.Port,
			Subject:   strconv.FormatInt(u.UserID, 10), //user id as string
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(Config.JwtSecret))
	if err != nil {
		response := NewResponse(http.StatusInternalServerError, "", nil)
		response.Err = err
		respondWithError(w, response)
		return
	}

	respondWithJSON(w, NewResponse(http.StatusOK, "", map[string]string{"token": signedToken}))
}
