package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mongodb/mongo-go-driver/mongo"
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

//UserHandler handles requests to /user/{id}
func UserHandler(w http.ResponseWriter, r *http.Request) {
	//get userid from path as integer
	vars := mux.Vars(r)
	useridStr := vars["id"]
	userid, err := strconv.Atoi(useridStr)
	if err != nil { //if userid can't be converted to integer, tell the client it's WRONG
		http.Error(w, "Invalid user id", 400)
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
		httpError(w, "Database Error", 500) // unexpected result
	}

	jsonResult, err := json.Marshal(result)
	if err != nil { //if can't marshal as json, panic with the error (500)
		httpError(w, "Database Error", 500) // can't determine whats wanted
	}

	//all good, tell the client json is incoming, and send it the json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(jsonResult)
}
