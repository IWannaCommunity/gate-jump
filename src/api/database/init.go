package database

import (
	"database/sql"
	"fmt"

	"github.com/IWannaCommunity/gate-jump/src/api/log"
	_ "github.com/go-sql-driver/mysql"
)

const version uint8 = 2

var db *sql.DB

func Connect(user, password, dbname string) {
	var err error
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@/%s?charset=utf8mb4&parseTime=True&interpolateParams=true", user, password, dbname))
	if err != nil {
		log.Fatal(err)
	}
}

func Init() error {
	err, exists := doesTableExist("meta")
	if err != nil {
		return err
	}

	// If the meta database does not exist, setup the first two schemas
	if exists == false {
		err := setupSchema("00001_inital.sql")
		if err != nil {
			return err
		}

		err = setupSchema("00002_meta.sql")
		if err != nil {
			return err
		}

		result, err := db.Exec(`INSERT INTO meta ( db_version ) VALUES ( ? );`, version)
		if err != nil {
			return err
		}

		log.Info(result.LastInsertId())
		log.Info(result.RowsAffected())

		return nil
	}

	// Run other migrations if required
	rows := db.QueryRow(`SELECT * FROM meta LIMIT 1`)
	current := *new(uint8)
	rows.Scan(&current) // TODO: error check this

	log.Debug("Reported db version", current)

	if current != version {
		//TODO: do something in the future when we have more than two migration schemas
	}

	return nil
}
