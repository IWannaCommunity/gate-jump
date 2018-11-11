package routers

import (
	"database/sql"
	"encoding/json"
	"gate-jump/src/api/util"
	"net/http"
	"strconv"
	"time"

	"github.com/IWannaCommunity/gate-jump/src/api/authentication"
	"github.com/IWannaCommunity/gate-jump/src/api/database"
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
func getAlive(w http.ResponseWriter, r *http.Request) {
	res.New(http.StatusOK).JSON(w)
}

// get
func getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid User ID").Error(w)
		return
	}

	u := database.User{ID: int64(id)}

	auth, response := getAuthLevel(r, &u)
	if response != nil {
		response.Error(w)
		return
	}

	if serr := u.GetUser(auth); serr.Err != nil {
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
func getUserByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	u := database.User{Name: &name}

	if serr := u.GetUser(authentication.PUBLIC); serr.Err != nil {
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
func getUsers(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 10 || count < 1 {
		count = 50
	}
	if start < 0 {
		start = 0
	}

	auth, response := getAuthLevel(r, nil)
	if response != nil {
		response.Error(w)
		return
	}

	users, serr := database.GetUsers(start, count, auth)
	if serr != nil {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}

	res.New(http.StatusOK).SetUsers(users).JSON(w)
}

// register
func createUser(w http.ResponseWriter, r *http.Request) {
	var u database.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil || u.Name == nil || u.Password == nil || u.Email == nil {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid Request Payload").Error(w)
		return
	}
	defer r.Body.Close()

	checkuser := u

	if util.IsNumeric(*checkuser.Name) { // dont allow
		res.New(http.StatusNoContent).SetErrorMessage("A Username Can't Be All Numbers").Error(w) // uncertain about 204 return
		return
	}

	if !util.IsValidEmail(*checkuser.Email) {
		res.New(http.StatusNoContent).SetErrorMessage("Invalid Email Format").Error(w)
		return
	}

	//check if user with name already exists; if not, we will get an ErrNoRows which is what we want
	if serr := checkuser.GetUserByName(authentication.SERVER); serr.Err == nil {
		res.New(http.StatusConflict).SetErrorMessage("User Already Exists").Error(w)
		return
	} else if serr.Err != sql.ErrNoRows {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}
	// check if user with email already exists; if not, we will get an ErrNoRows which is what we want
	if serr := checkuser.GetUserByEmail(authentication.SERVER); serr.Err == nil {
		res.New(http.StatusConflict).SetErrorMessage("Email Already In Use").Error(w)
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

	if serr := u.CreateUser(); serr != nil {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}

	res.New(http.StatusCreated).JSON(w)
}

// update
func updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid User ID").Error(w)
		return
	}

	var u database.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid Request Payload").Error(w)
		return
	}
	defer r.Body.Close()
	u.ID = int64(id) // set expected id to url id value

	// get auth level of the request for the given id
	auth, response := getAuthLevel(r, &u)
	if response != nil {
		response.Error(w)
		return
	}
	// api requests made by permissions less than users can't edit any other user so reject them completely
	if auth < authentication.USER {
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

	serr := u.UpdateUser(auth)
	if serr.Err != nil {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}

	res.New(http.StatusOK).SetUser(u).JSON(w)
}

// delete
func deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid User ID").Error(w)
		return
	}

	u := database.User{ID: int64(id)}

	auth, response := getAuthLevel(r, &u)
	if response != nil {
		response.Error(w)
		return
	}
	if auth < authentication.USER { // they arent the given user
		res.New(http.StatusUnauthorized).SetErrorMessage("Invalid Permissions").Error(w)
		return
	}
	if serr := u.DeleteUser(); serr.Err != nil {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}
	res.New(http.StatusAccepted).JSON(w)
}

// login
func validateUser(w http.ResponseWriter, r *http.Request) {
	var lr LoginRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&lr); err != nil {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid Request Payload").Error(w)
		return
	}
	defer r.Body.Close()

	var u database.User
	u.Name = &lr.Username

	//get the user; if no user by that name, return 401, if other error, 500
	if serr := u.GetUserByName(authentication.SERVER); serr.Err != nil {
		if serr.Err == sql.ErrNoRows {
			res.New(http.StatusUnauthorized).SetErrorMessage("User Doesn't Exist").Error(w)
			return
		} else if serr.Err != nil {
			res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
			return
		}
	}

	if u.Deleted != nil && *u.Deleted {
		if serr := u.UnflagDeletion(); serr.Err != nil {
			res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
			return
		}
	}

	//check the password
	if err := bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(lr.Password)); err != nil {
		res.New(http.StatusUnauthorized).SetErrorMessage("Wrong Password").Error(w)
		return
	}

	signedToken, err := u.CreateToken()
	if err != nil {
		res.New(http.StatusInternalServerError).SetErrorMessage("Failed Creating Token").Error(w)
		return
	}

	u.LastToken = &signedToken
	u.LastLogin = &[]time.Time{time.Now()}[0] // how to get pointer from function call (its gross): goo.gl/9BXtsj
	u.LastIP = &r.RemoteAddr
	if serr := u.UpdateUser(authentication.SERVER); serr.Err != nil {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}

	res.New(http.StatusOK).SetToken(signedToken).JSON(w)
}

func refreshUser(w http.ResponseWriter, r *http.Request) {
	// get user
	claims := r.Context().Value(authentication.CLAIMS).(authentication.Claims) // claims at this point are validated so refresh is allowed
	var u database.User
	u.ID = claims.ID
	if serr := u.GetUser(authentication.SERVER); serr.Err != nil {
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
	if serr := u.UpdateUser(authentication.SERVER); serr.Err != nil {
		res.New(http.StatusInternalServerError).SetInternalError(serr).Error(w)
		return
	}

	// return token
	res.New(http.StatusOK).SetToken(token).JSON(w)
}

// provide with request and said user and claims and confirm claims user exists and claims user's authentication level
//TODO: move this function to the authentication package so we can unexport ctx.Claims and ctx.Tokens
func getAuthLevel(r *http.Request, u1 *database.User) (authentication.Level, *res.Response) {
	ctx := r.Context().Value(authentication.CLAIMS).(authentication.Context) // confirmed valid on jwt layer

	if ctx.Claims.ID == 0 { // no claims exist
		return authentication.PUBLIC, nil
	}

	var u2 database.User
	u2.ID = ctx.Claims.ID
	serr := u2.GetUser(authentication.SERVER)
	if serr.Err == sql.ErrNoRows { // claims user wasn't found
		return authentication.PUBLIC, res.New(http.StatusUnauthorized).SetErrorMessage("Token's User Doesn't Exist")
	} else if serr.Err != nil {
		return authentication.PUBLIC, res.New(http.StatusInternalServerError).SetInternalError(serr)
	}
	if u2.LastToken == nil || ctx.Token != *u2.LastToken { // confirm the token is actually the last token used by the user
		return authentication.PUBLIC, res.New(http.StatusUnauthorized).SetErrorMessage("Token's User And Found User's Last Token Are Not The Same")
	}

	// we assume the username of the claimed user and the found user (u2) is the same because we searched by name
	if u2.Banned != nil && *u2.Banned { // fuck this guy in particular
		return authentication.PUBLIC, nil
	}

	if u1 == nil { // we aren't editing a user directly so no user was provided
		if u2.Admin != nil && *u2.Admin { // u2 is an admin
			return authentication.ADMIN, nil
		} else { // u2 is not an admin
			return authentication.PUBLIC, nil
		}
	} else {
		if u1.ID == u2.ID { // is u1 u2?
			if u2.Admin != nil && *u2.Admin { // is u2 an admin like they say they are?
				return authentication.ADMINUSER, nil
			} else { // u2 is not an admin but is u1
				return authentication.USER, nil
			}
		} else if u2.Admin != nil && *u2.Admin { // is u2 not u1 but is an admin?
			return authentication.ADMIN, nil
		}
		return authentication.PUBLIC, nil // u2 is neither u1 or an admin
	}
}
