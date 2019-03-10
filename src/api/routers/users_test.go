package routers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/IWannaCommunity/gate-jump/src/api/log"
	tst "github.com/IWannaCommunity/gate-jump/src/api/testing"
	"github.com/stretchr/testify/assert"
)

const tableCreationQuery = `CREATE TABLE users (
    id INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL,
    password CHAR(60) BINARY NOT NULL,
    email VARCHAR(100),
    country CHAR(2),
    locale VARCHAR(20),
    date_created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    admin BOOL NOT NULL DEFAULT FALSE,
    verified BOOL NOT NULL DEFAULT FALSE,
    banned BOOL NOT NULL DEFAULT FALSE,
    last_token BLOB,
    last_login DATETIME,
    last_ip VARCHAR(50),
    deleted BOOL NOT NULL DEFAULT FALSE,
    date_deleted DATETIME,
    PRIMARY KEY (id)
)`

var te *tst.TestingEnv

func TestMain(m *testing.M) {
	var err error

	/*
		err = settings.FromFile("config/config.json")
		if err != nil {
			log.Fatal(err) // clearly couldn't get database variables
		}*/

	database, err := sql.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8mb4&parseTime=True&interpolateParams=true", "root", "", "gatejump"))
	if err != nil {
		log.Fatal(err) // can't run tests if we can't initialize the database
	}

	go Serve("10421", "444") // run router on port

	for router != nil {
		log.Info("It's nil...")
	}

	te.Init(database, router, tableCreationQuery)

	code := m.Run() // run tests

	os.Exit(code) // we finished the tsts

}

func TestAlive(t *testing.T) {
	te.Prepare("GET", "/")
	r := te.Request(nil)

	if assert.NoError(t, r.Err, te.Expect(nil)) {

		assert.Equal(t, http.StatusOK, r.Code, te.Expect(http.StatusOK))
		assert.True(t, r.Response.Success, te.Expect(true))

		assert.Nil(t, r.Response.Error, te.Expect(nil))
		assert.Nil(t, r.Response.Token, te.Expect(nil))
		assert.Nil(t, r.Response.User, te.Expect(nil))
		assert.Nil(t, r.Response.UserList, te.Expect(nil))

	}

}
