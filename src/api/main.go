package main

import (
	"github.com/IWannaCommunity/gate-jump/src/api/log"
)

func main() {

	f := log.Init()
	defer f.Close()

	log.Info("Setting up environment...")
	log.Info("Loading Configuration")

	LoadConfig("config/config.json")

	log.Info("Initializing Database")
	s := Server{LogFile: f}
	s.Initialize(Config.Database.Username, Config.Database.Password, Config.Database.Dsn)

	log.Info("Initializing Routes")
	s.InitializeRoutes()

	log.Info("Starting the gate-jump server now! Ctrl+C to quit.")
	s.Run(Config.Port, Config.SslPort)
}
