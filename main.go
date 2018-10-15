package main

import "log"

func main() {
	log.Println("Welcome to gate-jump server! Setting up environment...")

	log.Println("Loading Configuration")
	LoadConfig("config/config.json")

	log.Println("Initializing Logging")
	initLog(Config.Major, Config.Patch, Config.Minor) // major,patch,minor

	log.Println("Initializing Database")
	s := Server{}
	s.Initialize(Config.Database.Username, Config.Database.Password, Config.Database.Dsn)

	log.Println("Starting the gate-jump server now! Ctrl+C to quit.")
	s.Run(Config.Port)
}
