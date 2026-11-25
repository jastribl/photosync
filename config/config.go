package config

import (
	"encoding/json"
	"log"
	"os"
)

// Config is the struct that holds photosync config info
type Config struct {
	TokenFileLocation string    `json:"token-file-location"`
	TokenDoneSignal   chan bool `json:"-"`

	// OAuth config
	ClientID     string   `json:"client-id"`
	ClientSecret string   `json:"client-secret"`
	Scopes       []string `json:"scopes"`
	AuthURL      string   `json:"auth-url"`
	TokenURL     string   `json:"token-url"`
	RedirectURL  string   `json:"redirect-url"`
}

var configCache *Config

// NewConfig gets a new Config
func NewConfig() *Config {
	if configCache == nil {
		configCache = new(Config)
		configFile, err := os.Open("config/config.json")
		if err != nil {
			log.Fatal(err)
		}
		defer configFile.Close()
		jsonParser := json.NewDecoder(configFile)
		jsonParser.Decode(configCache)

		// init channels
		configCache.TokenDoneSignal = make(chan bool)
	}
	return configCache
}
