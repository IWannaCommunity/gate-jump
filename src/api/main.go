package main

import (
	"github.com/IWannaCommunity/gate-jump/src/api/log"
	"github.com/IWannaCommunity/gate-jump/src/api/settings"
)

func main() {

	f := log.Init()
	defer f.Close()

	log.Info("Setting up environment...")
	log.Info("Loading Configuration")

	settings.FromFile("config/config.json")

	log.Info("Connecting Database")
	s := Server{LogFile: f}
	s.Initialize(settings.Config.Database.Username,
		settings.Config.Database.Password,
		settings.Config.Database.Dsn)

	log.Info("Initializing Routes")
	s.InitializeRoutes()

	log.Info("Starting the gate-jump server now! Ctrl+C to quit.")
	s.Run(Config.Port, Config.SslPort)
}
