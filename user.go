package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mongodb/mongo-go-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

//UserObject is a representation of the user object in the database
type UserObject struct {
	Userid      int
	Username    string
	Password    string `json:"-"` //don't share the password via json responses
	Email       string
	Country     string
	DateCreated time.Time `bson:"dateCreated"`
	Verified    bool
	Banned      bool
	LastToken   string    `bson:"lastToken"`
	LastLogin   time.Time `bson:"lastLogin"`
	LastIP      string    `bson:"lastIP" json:"-"` //don't share the IP via json responses
}

// GetUserByName gets a users details by name. for registering and what not
func GetUserByName(w http.ResponseWriter, username string) (*UserObject, error) {
	result := UserObject{}
	collection := client.Database("gatejump").Collection("users")
	log.Println("Collection Created")
	err := collection.FindOne(nil, map[string]string{"username": username}).Decode(&result)
	log.Println("FindOne Completed")
	if err == mongo.ErrNoDocuments { // no user found, return nil, nil
		return nil, nil
	} else if err != nil { // no user found and some error
		return nil, err
	}
	return &result, nil // user found, return it
}

// GetUserByEmail gets a users details by name. for registering and what not
func GetUserByEmail(w http.ResponseWriter, email string) (*UserObject, error) {
	var result UserObject
	collection := client.Database("gatejump").Collection("users")
	err := collection.FindOne(context.Background(), map[string]string{"email": email}).Decode(&result)
	if err == mongo.ErrNoDocuments { // no user found, return nil, nil
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &result, nil // user found, return it
}

// GetUser handles requests to /user/{id}
func GetUser(w http.ResponseWriter, r *http.Request) {
	//get userid from path as integer
	vars := mux.Vars(r)
	useridStr := vars["id"]
	userid, err := strconv.Atoi(useridStr)
	if err != nil { //if userid can't be converted to integer, tell the client it's WRONG
		httpResponse(w, "Invalid User ID", 400)
		return
	}

	//get user from database
	collection := client.Database("gatejump").Collection("users")
	result := UserObject{}
	err = collection.FindOne(context.Background(), map[string]int{"userid": userid}).Decode(&result)
	if err == mongo.ErrNoDocuments { //if no user found, return 404
		httpResponse(w, "User Not Found", 404)
		return
	} else if err != nil {
		httpError(w, "Database Error in GetUser", 500) // unexpected result
		return
	}

	jsonResult, err := json.Marshal(result)
	if err != nil { //if can't marshal as json, panic with the error (500)
		httpError(w, "Database Error in GetUser/jsonResult", 500) // can't determine whats wanted
		return
	}

	//all good, tell the client json is incoming, and send it the json
	httpResponseJson(w, 200, jsonResult)
	return
}

// Register registers a user into the database
func Register(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	email := r.FormValue("email")

	log.Println("Form Values Processed")

	// try to find a user with duplicate name
	getUser, err := GetUserByName(w, username)
	if err != nil {
		httpResponse(w, "Internal Error Getting User By Name", 500)
		return
	}
	if getUser != nil {
		httpResponse(w, "Username In Use", 409)
		return
	}

	log.Println("No Username In Use Found")

	// try to find a user with duplicate email
	getEmail, err := GetUserByEmail(w, email)
	if err != nil {
		httpResponse(w, "Internal Error Getting User By Email", 409)
		return
	}
	if getEmail != nil {
		httpResponse(w, "Email In Use", 409)
		return
	}

	log.Println("No Email In Use Found")

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		httpResponse(w, "Internal Error Encrypting Password", 500)
		return
	}

	log.Println("Password Hashed")

	var user UserObject

	user.Username = username
	user.Email = email
	user.Password = string(hashPassword)

	collection := client.Database("gatejump").Collection("users")
	result, err := collection.InsertOne(context.Background(), user)
	log.Println(result)
	log.Println(err)
	httpResponse(w, "Created", 201)
	return
}
