package main

import (
	"github.com/IWannaCommunity/gate-jump/src/api/database"
	"github.com/IWannaCommunity/gate-jump/src/api/log"
	"github.com/IWannaCommunity/gate-jump/src/api/routers"
	"github.com/IWannaCommunity/gate-jump/src/api/settings"
)

func main() {

	log.Init()
	defer log.Close()

	log.Info("Setting up environment...")
	log.Info("Loading Configuration")

	settings.FromFile("config/config.json")

	log.Info("Connecting to Database")

	database.Connect(settings.Database.Username,
		settings.Database.Password,
		settings.Database.Dsn)

	log.Info("Initalizing Database")

	err := database.Init()
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Setting Up Routes")

	routers.Serve(settings.Port, settings.SslPort)

	log.Info("The gate-jump server has started! Ctrl+C to quit.")

}
