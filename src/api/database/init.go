package database

import (
	"io"

	"github.com/spidernest-go/db/lib/sqlbuilder"
	"github.com/spidernest-go/db/mysql"
	log "github.com/spidernest-go/logger"
	"github.com/spidernest-go/migrate"
)

var sess sqlbuilder.Database

func Init(dsn, host, user, pass string) {
	var err error
	sess, err = mysql.Open(mysql.ConnectionURL{
		Database: dsn,
		User:     user,
		Password: pass,
	})
	if err != nil {
		log.Fatal().Msgf("failed initalizing the database: %v", err)
	}
}

func Apply(version uint8, name string, r io.Reader) error {
	return migrate.Apply(version, name, r, sess)
}
