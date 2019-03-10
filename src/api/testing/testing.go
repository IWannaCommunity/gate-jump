package testing

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
)

type Payload struct {
	Success  bool        `json:"success"`
	Error    *string     `json:"error,omitempty"`
	Token    *string     `json:"token,omitempty"`
	User     interface{} `json:"user,omitempty"`
	UserList interface{} `json:"userList,omitempty"`
}

type TestPayload struct {
	code    int
	payload *Payload
}

type TestingEnv struct {
	s             *sql.DB
	r             *mux.Router
	tp            *TestPayload
	method        string
	url           string
	creationQuery string
}

func (te *TestingEnv) Init(s *sql.DB, r *mux.Router, creationQuery string) *TestingEnv {
	te.s = s
	te.r = r
	te.creationQuery = creationQuery
	return te
}

func (te *TestingEnv) Prepare(method string, url string) {

}

func clearTable(db *sql.DB) {
	db.Exec("DELETE FROM users")
	db.Exec("ALTER TABLE users AUTO_INCREMENT = 1")
}

func ensureTableExists(db *sql.DB, creationQuery string) error {
	if _, err := db.Exec(creationQuery); err != nil {
		return err
	}
	return nil
}

func (te *TestingEnv) executeRequest(req []byte, method string, url string) (*TestingEnv, error) {

	// Ensure a clean database

	// Make API Request
	httpRequest, _ := http.NewRequest(method, url, nil)
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
