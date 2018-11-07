package settings

import (
	"encoding/json"
	"os"

	"github.com/IWannaCommunity/gate-jump/src/api/log"
)

// Config configuration global variable
var Config *Configuration

// DatabaseConfig database configuration information (maybe not needed)
type databaseConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Dsn      string `json:"dsn"`
}

// HTTPS config info - if missing, then should fallback to HTTP
type httpsConfig struct {
	CertFile string `json:"certFile"`
	KeyFile  string `json:"keyFile"`
}

// Configuration data structure containing relevant database information (basically hides jwt from github)
type Configuration struct {
	Database          databaseConfig `json:"database"`
	Https             httpsConfig    `json:"https"`
	RouteBase         string         `json:"routebase"`
	Port              string         `json:"port"`
	SslPort           string         `json:"sslPort"`
	Host              string         `json:"host"`
	Protocol          string         `json:"protocol"`
	JwtSecret         string         `json:"jwt_secret"`
	DiscordWebhookURL string         `json:"discord_webhook_url"`
	Major             int            `json:"major"`
	Patch             int            `json:"patch"`
	Minor             int            `json:"minor"`
}

// FromFile loads the config.json file into the configuration struct and set Config to it
func FromFile(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Fatal("Configuration file " + filename + " does not exist!")
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
