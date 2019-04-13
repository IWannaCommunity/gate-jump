package routers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/IWannaCommunity/gate-jump/src/api/authentication"
	"github.com/IWannaCommunity/gate-jump/src/api/database"
	"github.com/IWannaCommunity/gate-jump/src/api/log"
	"github.com/IWannaCommunity/gate-jump/src/api/mailer"
	"github.com/IWannaCommunity/gate-jump/src/api/res"
	"github.com/IWannaCommunity/gate-jump/src/api/settings"
	"github.com/IWannaCommunity/gate-jump/src/api/util"
	smtp "github.com/go-mail/mail"
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
	defer log.Bench(time.Now(), "api/v0", r.RemoteAddr, http.StatusOK)
	res.New(http.StatusOK).JSON(w)
}

// get
func getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	uuid := vars["uuid"]
	u := database.User{UUID: &uuid}

	auth, response := getAuthLevel(r, &u)
	if response != nil {
		response.Error(w)
		return
	}

	if serr := u.GetUser(auth); serr.Err == sql.ErrNoRows {
		res.New(http.StatusNotFound).SetErrorMessage("User Not Found").Error(w)
		return
	} else if serr.Err != nil {
		res.New(http.StatusInternalServerError).SetInternalError(&serr).Error(w)
		return

	}

	res.New(http.StatusOK).SetUser(u).JSON(w)
}

// get via name
func getUserByName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	u := database.User{Name: &name}

	if serr := u.GetUserByName(authentication.PUBLIC); serr.Err != nil {
		switch serr.Err {
		case sql.ErrNoRows:
			res.New(http.StatusNotFound).SetErrorMessage("User Not Found").Error(w)
		default:
			res.New(http.StatusInternalServerError).SetInternalError(&serr).Error(w)
		}
		return
	}

	res.New(http.StatusOK).SetUser(u).JSON(w)
}

// get multiple
func getUsers(w http.ResponseWriter, r *http.Request) {
	count, _ := strconv.Atoi(r.FormValue("count"))
	start, _ := strconv.Atoi(r.FormValue("start"))

	if count > 50 {
		count = 50
	} else if count < 0 {
		count = 0
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
	if serr.Err != nil {
		res.New(http.StatusInternalServerError).SetInternalError(&serr).Error(w)
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

	// Generate a random string that will be used for email verification in the background
	ml := *new(database.MagicLink)
	cherr := make(chan error)
	chstr := make(chan []byte)
	go util.CreateRandomString(32, 1, chstr, cherr)

	// Validate user input
	if !util.IsValidUsername(*checkuser.Name) || util.IsValidEmail(*checkuser.Name) {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid Username Format").Error(w)
		return
	}

	if !util.IsValidEmail(*checkuser.Email) {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid Email Format").Error(w)
		return
	}

	if !util.IsValidPassword(*checkuser.Password) {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid Password Format").Error(w)
		return
	}

	//check if user with name already exists; if not, we will get an ErrNoRows which is what we want
	if serr := checkuser.GetUserByName(authentication.SERVER); serr.Err == nil {
		res.New(http.StatusConflict).SetErrorMessage("Username Already Exists").Error(w)
		return
	} else if serr.Err != sql.ErrNoRows {
		res.New(http.StatusInternalServerError).SetInternalError(&serr).Error(w)
		return
	}
	// check if user with email already exists; if not, we will get an ErrNoRows which is what we want
	if serr := checkuser.GetUserByEmail(authentication.SERVER); serr.Err == nil {
		res.New(http.StatusConflict).SetErrorMessage("Email Already In Use").Error(w)
		return
	} else if serr.Err != sql.ErrNoRows {
		res.New(http.StatusInternalServerError).SetInternalError(&serr).Error(w)
		return
	}

	//hash the password
	hashpwd, err := bcrypt.GenerateFromPassword([]byte(*u.Password), 12)
	if err != nil {
		res.New(http.StatusInternalServerError).SetErrorMessage("Failed Encrypting Password").Error(w)
		return
	}
	*u.Password = string(hashpwd)

	if serr := u.CreateUser(); serr.Err != nil {
		res.New(http.StatusInternalServerError).SetInternalError(&serr).Error(w)
		return
	}

	// create a new magic link
	err = <-cherr
	if err != nil {
		// we can't "fail" here, so we will pass and ask them manually ask for a new magiclink
		log.Error("Could not generate a random magiclink string, %v", err)
	}

	// TODO: Probably don't need strings.Builder, maybe do string(<-chstr)
	buf := new(strings.Builder)
	buf.Write(<-chstr)
	ml.UserID = *u.ID
	ml.Magic = buf.String()
	if serr := ml.CreateMagicLink(); serr.Err != nil {
		// can't fail here either
		log.Error("Could not create a magiclink, %v", serr.Err)
	}
	log.Info("New Magiclink ID: ", ml.ID)

	res.New(http.StatusCreated).JSON(w)

	msg := smtp.NewMessage()
	msg.SetHeader("From", settings.Mailer.User)
	msg.SetHeader("To", *checkuser.Email)
	msg.SetHeader("Subject", "Account Verification for I Wanna Community")
	// TODO: Change the URL here, hardcoded for now...
	msg.SetBody("text/plain", `In order to complete account registration,
		please verify your email by clicking the link below.

		https://localhost:80/verify/`+ml.Magic)
	mailer.Outbox <- msg
}

// verifyUser verifies the user's account
func verifyUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	str := vars["magic"]

	ml := database.MagicLink{Magic: str}
	serr := ml.GetMagicLinkFromMagicString()

	if serr.Err != nil {
		res.New(http.StatusInternalServerError).SetInternalError(&serr).Error(w)
		return
	}

	usr := database.User{ID: &ml.UserID}
	serr = usr.GetUser(authentication.SERVER)

	if serr.Err != nil {
		res.New(http.StatusInternalServerError).SetInternalError(&serr).Error(w)
		return
	}

	// TODO: we need to get rid of pointer values in structs unless
	// absolutely necessary, that way we don't get monsters like this
	usr.Verified = &[]bool{true}[0]
	serr = usr.UpdateUser(authentication.SERVER)

	if serr.Err != nil {
		res.New(http.StatusInternalServerError).SetInternalError(&serr).Error(w)
		return
	}

	// delete the link from the database, we can fail here without panicing
	serr = ml.DeleteMagicLinkFromMagicString()

	if serr.Err != nil {
		log.Warning("Unable to delete a magic link from the database, %v", serr.Err)
	}

	// send a successful registration email
	msg := smtp.NewMessage()
	msg.SetHeader("From", settings.Mailer.User)
	msg.SetHeader("To", *usr.Email)
	msg.SetHeader("Subject", "Welcome to I Wanna Community!")
	msg.SetBody("text/plain", "Thank you for verifying and registering with I Wanna Community, we hope your stay with us is pleasant!")
	mailer.Outbox <- msg

	res.New(http.StatusAccepted).JSON(w)
}

// update
func updateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uuid := vars["uuid"]

	var u database.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&u); err != nil {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid Request Payload").Error(w)
		return
	}
	defer r.Body.Close()
	u.UUID = &uuid // set expected id to url id value

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
		res.New(http.StatusInternalServerError).SetInternalError(&serr).Error(w)
		return
	}

	res.New(http.StatusOK).SetUser(u).JSON(w)
}

// delete
func deleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	uuid := vars["id"]
	u := database.User{UUID: &uuid}

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
		res.New(http.StatusInternalServerError).SetInternalError(&serr).Error(w)
		return
	}
	res.New(http.StatusAccepted).JSON(w)
}

// login
func validateUser(w http.ResponseWriter, r *http.Request) {
	var lr LoginRequest

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&lr); err != nil || lr.Username == "" || lr.Password == "" {
		res.New(http.StatusBadRequest).SetErrorMessage("Invalid Request Payload").Error(w)
		return
	}
	defer r.Body.Close()

	var u database.User
	u.Name = &lr.Username

	//get the user; if no user by that name, return 401, if other error, 500
	if serr := u.GetUserByName(authentication.SERVER); serr.Err == sql.ErrNoRows {
		res.New(http.StatusUnauthorized).SetErrorMessage("User Doesn't Exist").Error(w)
		return
	} else if serr.Err != nil {
		res.New(http.StatusInternalServerError).SetInternalError(&serr).Error(w)
		return
	}

	// check if they are banned
	if u.Banned != nil && *u.Banned {
		res.New(http.StatusUnauthorized).SetErrorMessage("Account Banned").Error(w)
		return
	}

	//check the password
	if err := bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(lr.Password)); err != nil {
		res.New(http.StatusUnauthorized).SetErrorMessage("Wrong Password").Error(w)
		return
	}

	// check if they are deleted, if so undelete them
	if u.Deleted != nil && *u.Deleted {
		if serr := u.UnflagDeletion(); serr.Err != nil {
			res.New(http.StatusInternalServerError).SetInternalError(&serr).Error(w)
			return
		}
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
		res.New(http.StatusInternalServerError).SetInternalError(&serr).Error(w)
		return
	}

	res.New(http.StatusOK).SetToken(signedToken).JSON(w)
}

func refreshUser(w http.ResponseWriter, r *http.Request) {
	// get user
	ctx := r.Context().Value(authentication.CLAIMS).(authentication.Context) // claims at this point are validated so refresh is allowed
	claims := ctx.Claims

	var u database.User
	u.ID = claims.ID

	auth, response := getAuthLevel(r, &u)
	if response != nil {
		response.Error(w)
		return
	}

	if !(auth == authentication.USER || auth == authentication.ADMINUSER) {
		res.New(http.StatusUnauthorized).SetErrorMessage("Invalid Token Provided").Error(w)
		return
	}

	if serr := u.GetUser(authentication.SERVER); serr.Err != nil {
		res.New(http.StatusInternalServerError).SetInternalError(&serr).Error(w)
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
		res.New(http.StatusInternalServerError).SetInternalError(&serr).Error(w)
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
	u2.UUID = bearer.UUID
	serr := u2.GetUser(authentication.SERVER)
	if serr.Err == sql.ErrNoRows { // claims user wasn't found
		return authentication.PUBLIC, res.New(http.StatusUnauthorized).SetErrorMessage("Token's User Doesn't Exist")
	} else if serr.Err != nil {
		return authentication.PUBLIC, res.New(http.StatusInternalServerError).SetInternalError(&serr)
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
