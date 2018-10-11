package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
)

// Config configuration global variable
var Config *Configuration

// DatabaseConfig database configuration information (maybe not needed)
type DatabaseConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Dsn      string `json:"dsn"`
}

// Configuration data structure containing relevant database information (basically hides jwt from github)
type Configuration struct {
	Database          DatabaseConfig `json:"database"`
	RouteBase         string         `json:"routebase"`
	Port              int            `json:"port"`
	Host              string         `json:"host"`
	Protocol          string         `json:"protocol"`
	JwtSecret         string         `json:"jwt_secret"`
	MapImagePath      string         `json:"map_image_path"`
	DiscordWebhookURL string         `json:"discord_webhook_url"`
}

// PortStr helper function
func (conf *Configuration) PortStr() string {
	return strconv.Itoa(conf.Port)
}

// LoadConfig loads the config.json file into the configuration struct and set Config to it
func LoadConfig(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Panic("Configuration file " + filename + " does not exist!")
	}

	file, _ := os.Open(filename)
	decoder := json.NewDecoder(file)
	con := new(Configuration)
	err := decoder.Decode(&con)
	if err != nil {
		return err
	}
	Config = con
	return nil
}
