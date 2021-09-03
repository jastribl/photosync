package photos

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/jastribl/photosync/config"
	"golang.org/x/oauth2"
)

// Client holds all things for Photos requests
type Client struct {
	httpClient *http.Client
}

// HasToken returns if the user has a token
func HasToken(cfg *config.Config) bool {
	_, err := tokenFromFile(cfg)
	return err == nil
}

// Retrieves a token from a local file.
func tokenFromFile(cfg *config.Config) (*oauth2.Token, error) {
	f, err := os.Open(cfg.TokenFileLocation)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// SaveToken saves a token given a config
func SaveToken(cfg *config.Config, token *oauth2.Token) error {
	f, err := os.OpenFile(cfg.TokenFileLocation, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

// GetAuthConfig returns a new auth config
func GetAuthConfig(cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes:       cfg.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.AuthURL,
			TokenURL: cfg.TokenURL,
		},
		RedirectURL: cfg.RedirectURL,
	}

}

// NewClientForUser gets a new client for a user using the user token
func NewClientForUser(cfg *config.Config) (*Client, error) {
	tok, err := tokenFromFile(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		httpClient: GetAuthConfig(cfg).Client(context.Background(), tok),
	}, nil
}

// Get gets?
func (m *Client) Get(s string) (*http.Response, error) {
	return m.httpClient.Get(s)
}
