package database

import (
    "strings"
    "database/sql"

    "github.com/IWannaCommunity/gate-jump/src/api/settings"
    "github.com/IWannaCommunity/gate-jump/src/api/migrations"
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
    return err, true
}

func setupSchema(filename string) error {
    err, exists := doesTableExist("users")

    if err != nil {
        return err
    }

    if exists == false {
        f, err := migrations.ReadFile("src/migrations/")
        buf := new(strings.Builder)
        buf.Write(f)
        buf.Write(filename)

        stmt, err := db.Prepare(buf.String())

        if err != nil {
            return err
        }

        _, err = stmt.Exec()

        //TODO: Probably should return the result, and the error if it's not nil
        return err
    }

    //TODO: Probably should return an error about how there was nothing to do, instead of nil
    return nil
}
