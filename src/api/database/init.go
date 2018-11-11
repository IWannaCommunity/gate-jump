package database

import (
	"database/sql"
	"fmt"
	"gate-jump/src/api/routers"

	"github.com/IWannaCommunity/gate-jump/src/api/log"
	"github.com/IWannaCommunity/gate-jump/src/api/settings"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

var db *sql.DB

func InitServer() (*sql.DB, *mux.Router) {
	db := connect(settings.Database.Username,
		settings.Database.Password,
		settings.Database.Dsn)
	router := routers.Serve(settings.Port, settings.SslPort)
	return db, router
}

func connect(user, password, dbname string) *sql.DB {
	var err error
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8mb4&parseTime=True&interpolateParams=true", user, password, dbname))
	if err != nil {
		log.Fatal(err)
	}

	log.Debug(SetupUsers())
	return db
}
