package database

import (
	"database/sql"
	"fmt"

	"github.com/IWannaCommunity/gate-jump/src/api/log"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func Connect(user, password, dbname string) {
	var err error
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8mb4&parseTime=True&interpolateParams=true", user, password, dbname))
	if err != nil {
		log.Fatal(err)
	}

	log.Debug(setupSchema("00001_inital.sql"))
	log.Debug(setupSchema("00002_meta.sql"))
}
