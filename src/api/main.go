package main

import (
	"github.com/IWannaCommunity/gate-jump/src/api/log"
	"github.com/IWannaCommunity/gate-jump/src/api/settings"
	"github.com/IWannaCommunity/gate-jump/src/api/database"
)

func main() {

	f := log.Init()
	defer f.Close()

	log.Info("Setting up environment...")
	log.Info("Loading Configuration")

	settings.FromFile("config/config.json")

	log.Info("Connecting Database")

	database.Connect(settings.Database.Username,
		settings.Database.Password,
		settings.Database.Dsn)

	/*log.Info("Initializing Routes")
	s.InitializeRoutes()

	log.Info("Starting the gate-jump server now! Ctrl+C to quit.")
	s.Run(Config.Port, Config.SslPort)*/
}
