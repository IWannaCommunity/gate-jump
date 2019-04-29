package routers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/IWannaCommunity/gate-jump/src/api/authentication"
	"github.com/IWannaCommunity/gate-jump/src/api/database"
	"github.com/IWannaCommunity/gate-jump/src/api/log"
	"github.com/IWannaCommunity/gate-jump/src/api/mailer"
	"github.com/IWannaCommunity/gate-jump/src/api/res"
	"github.com/IWannaCommunity/gate-jump/src/api/settings"
	"github.com/IWannaCommunity/gate-jump/src/api/util"
	smtp "github.com/go-mail/mail"
	"golang.org/x/crypto/bcrypt"
)

// register
func createOwner(w http.ResponseWriter, r *http.Request) {

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
