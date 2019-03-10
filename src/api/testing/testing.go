package testing

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
)

// payload response returned from api. this should be passed in as an argument eventually
type Payload struct {
	Success  bool        `json:"success"`
	Error    *string     `json:"error,omitempty"`
	Token    *string     `json:"token,omitempty"`
	User     interface{} `json:"user,omitempty"`
	UserList interface{} `json:"userList,omitempty"`
}

// test payload containing information about the request. should include response time
type TestPayload struct {
	code    int
	payload *Payload
}

// includes everything relevant to making requests to the api directly through the router
type TestingEnv struct {
	s             *sql.DB
	r             *mux.Router
	tp            *TestPayload
	method        string
	url           string
	creationQuery string
}

// should initalize the database on its own eventually by being passed in a string
func (te *TestingEnv) Init(s *sql.DB, r *mux.Router, creationQuery string) *TestingEnv {
	te.s = s
	te.r = r
	te.creationQuery = creationQuery
	return te
}

func (te *TestingEnv) Prepare(method string, url string) *TestingEnv {
	// clean database for new setup
	_ = ensureTableExists(te.s, te.creationQuery)
	clearTable(te.s)

	// set method and url for api requests
	te.method = method
	te.url = url
	return te
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

func (te *TestingEnv) executeRequest(req []byte) (*TestingEnv, error) {
	// Make API Request
	httpRequest, _ := http.NewRequest(te.method, te.url, nil)
	httpTestRecorder := httptest.NewRecorder()
	te.r.ServeHTTP(httpTestRecorder, httpRequest)
	tp := TestPayload{}
	tp.code = httpTestRecorder.Code
	err := json.NewDecoder(httpTestRecorder.Body).Decode(tp.payload)
	if err != nil { // unmarshal failed somehow
		return te, err
	}
	return te, nil
}
