package database

import (
	"database/sql"
	"fmt"

	"github.com/IWannaCommunity/gate-jump/src/api/log"
	_ "github.com/go-sql-driver/mysql"
)

const version uint8 = 11

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

		result, err := db.Exec(`INSERT INTO meta ( db_version ) VALUES ( 2 );`)
		if err != nil {
			return err
		}

		log.Info(result.LastInsertId())
		log.Info(result.RowsAffected())
	}

	// Run other migrations if required
	rows := db.QueryRow(`SELECT * FROM meta LIMIT 1`)
	current := *new(uint8)
	rows.Scan(&current) // TODO: error check this

	log.Debug("Database Schema reported as Version", current)

	if current != version {
		switch current {

		case 2:
			log.Info("Migrating current Database Schema to 3")
			err := setupSchema("00003_magiclinks.sql")
			if err != nil {
				log.Error("Schema failed to execute successfully.")
				return err
			}
			fallthrough

		case 3:
			log.Info("Migrating current Database Schema to 4")
			err := setupSchema("00004_uuid.sql")
			if err != nil {
				log.Error("Schema failed to execute successfully.")
				return err
			}
			fallthrough

		case 4:
			log.Info("Migrating current Database Schema to 5")
			err := setupSchema("00005_scopes.sql")
			if err != nil {
				log.Error("Schema failed to execute successfully.")
				return err
			}
			fallthrough

		case 5:
			log.Info("Migrating current Database Schema to 6")
			err := setupSchema("00006_groups.sql")
			if err != nil {
				log.Error("Schema failed to execute successfully")
				return err
			}
			fallthrough

		case 6:
			log.Info("Migrating current Database Schema to 7")
			err := setupSchema("00007_permissions.sql")
			if err != nil {
				log.Error("Schema failed to execute successfully")
				return err
			}
			fallthrough

		case 7:
			log.Info("Migrating current Database Schema to 8")
			err := setupSchema("00008_memberships.sql")
			if err != nil {
				log.Error("Schema failed to execute successfully")
				return err
			}
			fallthrough

		case 8:
			log.Info("Migrate current Database Schema to 9")
			err := setupSchema("00009_logins.sql")
			if err != nil {
				log.Error("Schema failed to execute successfully")
				return err
			}
			fallthrough

		case 9:
			log.Info("Migrate current Database Schema to 10")
			err := setupSchema("00010_ipforlogins.sql")
			if err != nil {
				log.Error("Schema failed to execute successfully")
				return err
			}
			fallthrough

		case 10:
			log.Info("Migrate current Database Schema to 11")
			err := setupSchema("00011_epochforlogins.sql")
			if err != nil {
				log.Error("Schema failed to execute successfully")
				return err
			}
			fallthrough

		default:
			db.Exec(`UPDATE meta SET db_version=? WHERE db_version=?`, version, current)

			log.Info("No more Database Schema upgrades found, successfully upgraded to Database Schema version " + string(version) + ".")

		}
	}

	return nil
}
