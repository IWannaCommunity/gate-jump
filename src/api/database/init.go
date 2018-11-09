package database

import (
    "fmt"
    "database/sql"

    _ "github.com/go-sql-driver/mysql"
    "github.com/IWannaCommunity/gate-jump/src/api/log"
)

var db *sql.DB

func Connect(user, password, dbname string) {
	var err error
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8mb4&parseTime=True&interpolateParams=true", user, password, dbname))
	if err != nil {
		log.Fatal(err)
	}
}
