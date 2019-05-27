package main

import (
	"fmt"
	"os"

	"github.com/IWannaCommunity/gate-jump/src/api/routers"
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
	SetupLogging()      // Check for debug mode and apply settings appropriately,
	BuildInfo()         // Print build info to log.
	LoadConfiguration() // Load the configuration data into the server.
	// StartMailer() // Begin running the mailer service.
	// ConnectDatabase() // Connect to the database.
	// VerifySchemas() // Verify that the database schemas are correct.
	ListenAndServe() // Listen and serve API requests.
}

// CheckDebug checks if the argument "debug" for the first
func SetupLogging() {
	if len(os.Args) > 1 && os.Args[1] == "debug" {
		log.Stdout()
		log.Info().Msg("Server started in debugging mode. Logging exclusively to stdout...")
	} else {
		// log.Configure(...)
	}
}

func BuildInfo() {
	log.Info().Msgf("Starting passport: %s", fmt.Sprintf("%s.%s.%s %s, commit: %s, compiled @ %s", Minor, Patch, Major, Extra, Commit, Date))
}

func LoadConfiguration() {
	log.Info().Msg("Loading configuration data...")
	err := settings.FromFile("config/config.json")
	if err != nil {
		log.Fatal().Msgf("Failed loading configuration data: %v", err)
	}
}

func StartMailer() {
	log.Info().Msg("Initalizing mailer...")
	/*
		err := mailer.SMTPInit()
		if err != nil {
			log.Fatal().Msgf("Failed initalizing mailer: %v", err)
		}
		go mailer.Daemon()
	*/
}

func ConnectDatabase() {
	log.Info().Msg("Initalizing database...")
	/*
		database.Connect(
			settings.Database.Username,
			settings.Database.Password,
			settings.Database.Dsn)
	*/
}

func VerifySchemas() {
	log.Info().Msg("Verifying database schema...")
	/*
		err = database.Init()
		if err != nil {
			log.Fatal().Msgf("Database could not initalize: %v", err)
		}
	*/
}

func ListenAndServe() {
	log.Info().Msgf("Router listening at %s:%s", settings.Host, settings.Port)
	routers.Serve(settings.APIVer, settings.Port, settings.SslPort)
}
