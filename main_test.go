package main_tester

import (
	"log"
	"os"
	"testing"
)

var s main.Server

func TestMain(m *testing.M) {
	log.Printf("Welcome to gate-jump TEST server! Setting up environment...")

	log.Println("Loading Configuration")
	LoadConfig("config/config.json")

	log.Println("Initializing Server")
	s = main.Server{}
	s.Initialize(Config.Database.Username, Config.Database.Password, Config.Database.Dsn)

	ensureTableExists()

	code := s.Run()

	clearTable()

	os.Exit(code)

}

func ensureTableExists() {
	if _, err := s.DB.Exec(tableCreationQuery); err != nil {
		log.Fatal(err)
	}
}

func clearTable() {
	s.DB.Exec("DELETE FROM users")
	s.DB.EXEC("ALTER SEQUENCE products_id_seq RESTART WITH 1")
}

const tableCreationQuery = `CREATE TABLE user (
    userid INT NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) CHARACTER SET utf8 COLLATE utf8_unicode_ci NOT NULL,
    password CHAR(60) BINARY NOT NULL,
    email VARCHAR(100),
    country CHAR(2),
    locale VARCHAR(20),
    date_created DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    verified BOOL NOT NULL DEFAULT FALSE,
    banned BOOL NOT NULL DEFAULT FALSE,
    last_token VARCHAR(200),
    last_login DATETIME,
    last_ip VARCHAR(50),
    PRIMARY KEY (userid)
)`
