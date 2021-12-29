package config

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
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
	FreeBeforeDate                string           `json:"free-before-date"`
	RootPicturesDir               string           `json:"root-pictures-dir"`
	PicturePathSubstringsToIgnore []string         `json:"picture-path-substrings-to-ignore"`
	PicturePathRegexsToIgnore     []*regexp.Regexp `json:"-"`
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

		// Data Prepping
		configCache.PicturePathRegexsToIgnore = []*regexp.Regexp{}
		for _, regexToIgnore := range configCache.PicturePathSubstringsToIgnore {
			configCache.PicturePathRegexsToIgnore = append(
				configCache.PicturePathRegexsToIgnore,
				regexp.MustCompile(regexToIgnore),
			)
		}
	}
	return configCache
}
