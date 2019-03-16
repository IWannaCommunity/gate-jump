package testing

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
)

// payload response returned from api. this should be passed in as an argument eventually
type Response struct {
	Success  bool        `json:"success"`
	Error    *string     `json:"error,omitempty"`
	Token    *string     `json:"token,omitempty"`
	User     interface{} `json:"user,omitempty"`
	UserList interface{} `json:"userList,omitempty"`
}

// test payload containing information about the request. should include response time
type TestPayload struct {
	Code     int
	Err      error
	Response *Response
}

// includes everything relevant to making requests to the api directly through the router
type TestingEnv struct {
	s             *sql.DB
	r             *mux.Router
	lastRequest   interface{}
	method        string
	url           string
	creationQuery string
}

func (te *TestingEnv) ExpectedPayload(code int, err error, token string, user interface{}, userList interface{}) {

}

// should initalize the database on its own eventually by being passed in a string
func (te *TestingEnv) Init(s *sql.DB, r *mux.Router, creationQuery string) {
	te.s = s
	te.r = r
	te.creationQuery = creationQuery
}

func (te *TestingEnv) Prepare(method string, url string) {
	// clean database for new setup
	_ = ensureTableExists(te.s, te.creationQuery)
	clearTable(te.s)

	// set method and url for api requests
	te.method = method
	te.url = url
}

func clearTable(db *sql.DB) {
	db.Exec("DELETE FROM users")
	db.Exec("ALTER TABLE users AUTO_INCREMENT = 1")
}

func ensureTableExists(db *sql.DB, creationQuery string) error {
	if _, err := db.Exec(creationQuery); err != nil {
		return err // do we care?
	}
	return nil
}

func (te *TestingEnv) Request(clear bool, jsonRequest []byte) TestPayload {
	if clear {
		clearTable(te.s)
	}
	// Make API Request
	te.lastRequest = jsonRequest
	httpRequest, _ := http.NewRequest(te.method, te.url, bytes.NewBuffer(jsonRequest))
	httpTestRecorder := httptest.NewRecorder()
	te.r.ServeHTTP(httpTestRecorder, httpRequest)
	tp := TestPayload{}
	tp.Err = nil
	tp.Code = httpTestRecorder.Code
	temp, _ := ioutil.ReadAll(httpTestRecorder.Body)
	err := json.Unmarshal(temp, &tp.Response)
	if err != nil { // unmarshal failed somehow
		tp.Err = err
	}
	return tp
}

func (te *TestingEnv) Expect() string {
	return fmt.Sprintf("%s @ %s\twith \"%v\"", te.method, te.url, te.lastRequest)
}
