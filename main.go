package main

import (
	"log"
	"os"
)

func main() {

	f := initLog(0, 0, 0) // major,patch,minor
	if f != os.Stdout {
		defer func() {
			f.Close()
		}()
	}

	log.Println("Welcome to the gate-jump server! Setting up environment...")

	log.Println("Loading Configuration")
	LoadConfig("config/config.json")

	log.Println("Initializing Database")
	s := Server{LogFile: f}
	s.Initialize(Config.Database.Username, Config.Database.Password, Config.Database.Dsn)

	log.Println("Initializing Routes")
	s.InitializeRoutes()

	log.Println("Starting the gate-jump server now! Ctrl+C to quit.")
	s.Run(Config.Port, Config.SslPort)
}
