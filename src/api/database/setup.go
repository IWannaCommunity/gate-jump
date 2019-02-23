package database

import (
	"database/sql"
	"strings"

	"github.com/IWannaCommunity/gate-jump/src/api/migrations"
	"github.com/IWannaCommunity/gate-jump/src/api/settings"
)

func doesTableExist(name string) (error, bool) {
	stmt, err := db.Prepare(`SELECT *
        FROM information_schema.tables
        WHERE table_schema = ?
            AND table_name = ?
        LIMIT 1;`)

	if err != nil {
		return err, true
	}

	rows := stmt.QueryRow(settings.Database.Dsn, name)

	table := make(map[string]interface{})
	err = rows.Scan(&table)

	if err == sql.ErrNoRows {
		return nil, false
	}

	//TODO: this should be it's own error through errors.New but I'm lazy
	return nil, true
}

func setupSchema(filename string) error {
	f, err := migrations.ReadFile(filename)

	if err != nil {
		return err
	}

	buf := new(strings.Builder)
	buf.Write(f)

	stmt, err := db.Prepare(buf.String())

	if err != nil {
		return err
	}

	_, err = stmt.Exec()

	if err != nil {
		return err
	}

	//TODO: Probably should return the result, and the error if it's not nil
	return nil
}
