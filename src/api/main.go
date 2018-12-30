package main

import (
	"github.com/IWannaCommunity/gate-jump/src/api/database"
	"github.com/IWannaCommunity/gate-jump/src/api/log"
	"github.com/IWannaCommunity/gate-jump/src/api/mailer"
	"github.com/IWannaCommunity/gate-jump/src/api/routers"
	"github.com/IWannaCommunity/gate-jump/src/api/settings"
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

	// Logger Initialization
	log.Init()
	defer log.Close()

	info, err := BuildInfo()
	if err != nil {
		log.Wtf("Could not concat Build Information string, %v", err)
	}
	log.Verbose("Starting Gatejump v" + info)

	// Configuration Initialization
	log.Info("Parsing Configuration...")

	err = settings.FromFile("config/config.json")
	if err != nil {
		log.Fatal("Failed loading configuration: ", err)
	}

	log.Info("Configuration loaded!")

	// Mailer Initialization
	log.Info("Connecting to Mail Servers...")

	err = mailer.SMTPInit()
	if err != nil {
		log.Fatal("Could not communicate with local SMTP host, %v", err)
	}

	log.Info("Successfully connected to local SMTP host!")

	// Database Initialization
	log.Info("Attaching to Database...")

	database.Connect(settings.Database.Username,
		settings.Database.Password,
		settings.Database.Dsn)

	log.Info("Checking if the Database Schema is out of date...")

	err = database.Init()
	if err != nil {
		log.Fatal("Failed Initializing Database: ", err)
	}

	log.Info("Database is successfully attached!")

	// HTTP Initialization
	log.Info("Serving API Routes at " + settings.Host + ":" + settings.Port)

	routers.Serve(settings.Port, settings.SslPort)
}
