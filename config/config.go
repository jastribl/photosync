package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config is the struct that holds splitwise config info
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

	// Program Config
	FreeBeforeDate                string          `json:"free-before-date"`
	RootPicturesDir               string          `json:"root-pictures-dir"`
	PicturePathSubstringsToIgnore []string        `json:"picture-path-substrings-to-ignore"`
	FileNamesToIgnoreArray        []string        `json:"file-names-to-ignore"`
	FileNamesToIgnoreMap          map[string]bool `json:"-"`
}

var configCache *Config

// NewConfig gets a new Config
func NewConfig() *Config {
	if configCache == nil {
		configCache = new(Config)
		configFile, err := os.Open("config/config.json")
		defer configFile.Close()
		if err != nil {
			fmt.Println(err.Error())
		}
		jsonParser := json.NewDecoder(configFile)
		jsonParser.Decode(configCache)

		// init channels
		configCache.TokenDoneSignal = make(chan bool)

		// Data Prepping
		configCache.FileNamesToIgnoreMap = make(map[string]bool)
		for _, item := range configCache.FileNamesToIgnoreArray {
			configCache.FileNamesToIgnoreMap[item] = true
		}
	}
	return configCache
}
