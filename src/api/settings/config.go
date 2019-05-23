package settings

import (
	"encoding/json"
	"os"

	log "github.com/spidernest-go/logger"
)

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

type mailerConfig struct {
	Host string `json:"host"`
	Port string `json:"port"`
	User string `json:"user"`
	Pass string `json:"pass"`
}

type superuserConfig struct {
	Password string `json:"password"`
}

// Configuration and Settings
var (
	Database          databaseConfig
	Https             httpsConfig
	Mailer            mailerConfig
	SuperUser         superuserConfig
	RouteBase         string
	Port              string
	SslPort           string
	Host              string
	Protocol          string
	JwtSecret         string
	DiscordWebhookURL string
	APIVer            string
	Major             int
	Patch             int
	Minor             int
)

// FromFile loads the config.json file into Settings variables
func FromFile(filename string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Fatal().Msgf("Configuration file %s does not exist!", filename)
	}

	file, _ := os.Open(filename)
	decoder := json.NewDecoder(file)
	config := make(map[string]interface{})
	err := decoder.Decode(&config)

	if err != nil {
		return err
	}

	setActiveConfig(config)

	return nil
}

func setActiveConfig(configmap map[string]interface{}) {
	Database = databaseConfig{
		Username: configmap["database"].(map[string]interface{})["username"].(string),
		Password: configmap["database"].(map[string]interface{})["password"].(string),
		Dsn:      configmap["database"].(map[string]interface{})["dsn"].(string),
	}

	Https = httpsConfig{
		CertFile: configmap["https"].(map[string]interface{})["certFile"].(string),
		KeyFile:  configmap["https"].(map[string]interface{})["keyFile"].(string),
	}

	Mailer = mailerConfig{
		Host: configmap["mailer"].(map[string]interface{})["host"].(string),
		Port: configmap["mailer"].(map[string]interface{})["port"].(string),
		User: configmap["mailer"].(map[string]interface{})["user"].(string),
		Pass: configmap["mailer"].(map[string]interface{})["pass"].(string),
	}

	SuperUser = superuserConfig{
		Password: configmap["superuser"].(map[string]interface{})["password"].(string),
	}

	Host = configmap["host"].(string)
	Port = configmap["port"].(string)
	SslPort = configmap["sslPort"].(string)
	APIVer = configmap["apiVer"].(string)
}
