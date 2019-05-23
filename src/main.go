package main

import (
	"fmt"
	"os"

	"github.com/IWannaCommunity/gate-jump/src/api/settings"
	log "github.com/spidernest-go/logger"
)

var (
	Minor  string
	Patch  string
	Major  string
	Extra  string
	Commit string
	Date   string
)

func main() {
	var err error
	// Check if we are in debugging mode.
	if len(os.Args) > 1 && os.Args[1] == "debug" {
		log.Stdout()
		log.Info().Msg("Server started in debugging mode. Logging exclusively to stdout...")
	} else {
		// log.Configure(...)
	}

	// Compile build info.
	log.Info().Msgf("Starting passport: %s", fmt.Sprintf("%s.%s.%s %s, commit: %s, compiled @ %s", Minor, Patch, Major, Extra, Commit, Date))

	// Load configuration data.
	log.Info().Msg("Loading configuration data...")
	err = settings.FromFile("config/config.json")
	if err != nil {
		log.Fatal().Msgf("Failed loading configuration data: %v", err)
	}

	// Initalize the mailer.
	/*
		log.Info().Msg("Initalizing mailer...")
			err = mailer.SMTPInit()
			if err != nil {
				log.Fatal().Msgf("Failed initalizing mailer: %v", err)
			}
			go mailer.Daemon()
	*/

	// Initalize the database.
	/*
		log.Info().Msg("Initalizing database...")
		database.Connect(
			settings.Database.Username,
			settings.Database.Password,
			settings.Database.Dsn)
	*/

	// Verify database schema.
	/*
		log.Info().Msg("Verifying database schema...")
		err = database.Init()
		if err != nil {
			log.Fatal().Msgf("Database could not initalize: %v", err)
		}
	*/

	// Preparing to serve.
	log.Info().Msgf("Router listening at %s:%s", settings.Host, settings.Port)
	routers.Serve(settings.Port, settings.SslPort)

}
