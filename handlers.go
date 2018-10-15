package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/klazen108/sfm-server-go/config"
)

func (s *Server) getAlive(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]string{"alive": "true"})
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	u := User{UserID: int64(id)}
	if err := u.getUser(s.DB); err != nil {
		switch err {
		case sql.ErrNoRows:
			respondWithError(w, http.StatusNotFound, "User not found")
		default:
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, u)
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
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, users)
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request) {
	var u User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	//check if user with name already exists; if so, we will get an ErrNoRows which is what we want
	checkuser := u
	err := checkuser.getUserByName(s.DB)
	if err == nil {
		respondWithError(w, http.StatusBadRequest, "User already exists")
		return
	} else if err != sql.ErrNoRows {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	//hash the password
	hashpwd, err := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	u.Password = string(hashpwd)

	if err := u.createUser(s.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, u)
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var u User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid resquest payload")
		return
	}
	defer r.Body.Close()
	u.UserID = int64(id)

	if err := u.updateUser(s.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, u)
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	u := User{UserID: int64(id)}
	if err := u.deleteUser(s.DB); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"result": "success"})
}

// LoginRequest is the request expected on /login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	Admin    bool   `json:"admin"`
	jwt.StandardClaims
}

func (s *Server) validateUser(w http.ResponseWriter, r *http.Request) {
	var lr LoginRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&lr); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	defer r.Body.Close()

	var u User
	u.Username = lr.Username

	//get the user; if no user by that name, return 401, if other error, 500
	if err := u.getUserByName(s.DB); err != nil {
		if err == sql.ErrNoRows {
			respondWithError(w, http.StatusUnauthorized, "Invalid Account")
			return
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	//check the password
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(lr.Password)); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid Account")
		return
	}

	//create and sign the token
	claims := Claims{
		u.Username,
		u.Admin,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(), //expire in one hour
			Issuer:    config.Config.Host + ":" + config.Config.PortStr(),
			Subject:   strconv.FormatInt(u.UserID, 10), //user id as string
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(config.Config.JwtSecret))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"token": signedToken})
}
