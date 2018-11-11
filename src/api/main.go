package main

import (
	"github.com/IWannaCommunity/gate-jump/src/api/database"
	"github.com/IWannaCommunity/gate-jump/src/api/log"
	"github.com/IWannaCommunity/gate-jump/src/api/settings"
)

func main() {

	log.Init()
	defer log.Close()

	log.Info("Setting up environment...")
	log.Info("Loading Configuration")

	settings.FromFile("config/config.json")

	log.Info("Initializing Database and Router")

	_, _ = database.InitServer()

	log.Info("The gate-jump server has started! Ctrl+C to quit.")

}
