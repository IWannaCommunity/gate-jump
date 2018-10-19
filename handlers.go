package main

import (
	"database/sql"
	"encoding/json"
	"gate-jump/res"
	"net/http"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) getAlive(w http.ResponseWriter, r *http.Request) {
	res.New(http.StatusOK).SetData(map[string]bool{"alive": true}).JSON(w)
}

func (s *Server) getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid User ID").Error(w)
		return
	}

	u := User{ID: int64(id)}
	if serr := u.getUser(s.DB); err != nil {
		switch serr.Err {
		case sql.ErrNoRows:
			res.New(http.StatusNotFound).SetErrorMessage("User Not Found").Error(w)
		default:
			res.New(http.StatusInternalServerError).Error(w)
		}
		return
	}
	res.New(http.StatusOK).SetData(u).JSON(w)
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

	users, serr := getUsers(s.DB, start, count)
	if serr != nil {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}

	res.New(http.StatusOK).SetData(users).JSON(w)
}

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
	serr := checkuser.getUserByName(s.DB)
	if serr.Err == nil {
		res.New(http.StatusBadRequest).SetErrorMessage("User Already Exists").Error(w)
		return
	} else if serr.Err != sql.ErrNoRows {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}

	//hash the password
	hashpwd, err := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
	if err != nil {
		res.New(http.StatusInternalServerError).SetErrorMessage("Failed Encrypting Password").Error(w)
		return
	}
	u.Password = string(hashpwd)

	if serr := u.createUser(s.DB); serr != nil {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}
	res.New(http.StatusCreated).SetData(u).JSON(w)
}

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
	u.ID = int64(id)

	if serr := u.updateUser(s.DB); serr != nil {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}
	res.New(http.StatusOK).SetData(u).JSON(w)
}

func (s *Server) deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid User ID").Error(w)
		return
	}

	u := User{ID: int64(id)}
	if serr := u.deleteUser(s.DB); serr != nil {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}

	res.New(http.StatusOK).JSON(w)
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
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid Request Payload").Error(w)
		return
	}
	defer r.Body.Close()

	var u User
	u.Name = lr.Username

	//get the user; if no user by that name, return 401, if other error, 500
	if serr := u.getUserByName(s.DB); serr != nil {
		if serr.Err == sql.ErrNoRows {
			res.New(http.StatusUnauthorized).SetErrorMessage("Invalid Account").Error(w)
			return
		} else if serr.Err != nil {
			res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
			return
		}
	}
	//check the password
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(lr.Password)); err != nil {
		res.New(http.StatusInternalServerError).SetErrorMessage("Failed Decrypting Password").Error(w)
		return
	}

	//create and sign the token
	claims := Claims{
		u.Name,
		u.Admin,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(), //expire in one hour
			Issuer:    Config.Host + ":" + Config.Port,
			Subject:   strconv.FormatInt(u.ID, 10), //user id as string
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(Config.JwtSecret))
	if err != nil {
		res.New(http.StatusBadRequest).SetErrorMessage("Failed Creating Token").Error(w)
		return
	}
	res.New(http.StatusOK).SetToken(signedToken).JSON(w)
}
